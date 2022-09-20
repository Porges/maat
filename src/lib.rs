//! # 𓁦 maat 𓆄
//!
//! ## About
//!
//! `maat` is a library for randomized execution of property-based tests.
//!
//!
//! ```rust
//! #[test]
//! pub fn run() {
//!     property(|maat| {
//!         let x = maat.generate("x", i64(0, 100));
//!         let y = maat.generate("y", i64(0, 100));
//!         x + y == x + x || x < 10
//!     });
//! }
//! ```
//!
//! Output:
//! ```text
//! ---- test::run stdout ----
//! thread 'test::run' panicked at '
//! [maat] Falsified property with values:
//! x: i64 = 10
//! y: i64 = 0
//!
//! [maat] Original failing values were:
//! x: i64 = 12
//! y: i64 = 36
//!
//! ', src/lib.rs:287:13
//! ```
//!
//! ## Internals
//!
//! `maat` works in three modes:
//! - First, it runs in [Mode::Testing], where generation of values is very fast.
//!   This mode is used almost all the time, since test failure is going to be rare.
//! - Secondly, if it finds a failure case, it produces a recording for that failure
//!   by re-running the test with the same RNG state, but in [Mode::Recording].
//! - Finally, once it has a recording, it tries to shrink the recording by
//!   re-running the test in [Mode::Shrinking].

use dynamic::Dynamic;
use rand::SeedableRng;
use rand_xoshiro::Xoshiro256PlusPlus as RNG;
use std::{
    any::type_name,
    cell::RefCell,
    fmt::{Debug, Display, Write},
    ops::DerefMut,
    rc::Rc,
    time::Instant,
};

pub mod generators;

#[derive(Clone)]
pub struct Shrinkable<T> {
    value: T,
    shrink: Shrinker<T>,
}

type Shrinker<T> = Rc<dyn Fn(&T, &mut dyn FnMut(T) -> bool) -> bool>;

impl<T> Shrinkable<T> {
    fn shrink(&self, is_valid: &mut dyn FnMut(Shrinkable<T>) -> bool) -> bool {
        (self.shrink)(&self.value, &mut |v| {
            is_valid(Shrinkable {
                value: v,
                shrink: self.shrink.clone(),
            })
        })
    }
}

pub trait Generator<T> {
    /// This is the fast path, used during [Mode::Testing].
    fn generate(&self, rng: &mut dyn rand::RngCore) -> T;

    /// This is the slower path, used during [Mode::Recording].
    /// The returned [Shrinkable] is used during [Mode::Shrinking].
    fn generate_shrinkable(&self, rng: &mut dyn rand::RngCore) -> Shrinkable<T>;
}

// Why is this needed?
impl<Gen, T> Generator<T> for &Gen
where
    Gen: Generator<T> + ?Sized,
{
    fn generate(&self, rng: &mut dyn rand::RngCore) -> T {
        (*self).generate(rng)
    }

    fn generate_shrinkable(&self, rng: &mut dyn rand::RngCore) -> Shrinkable<T> {
        (*self).generate_shrinkable(rng)
    }
}

struct Generated<T> {
    name: &'static str,
    value: RefCell<Shrinkable<T>>,
}

impl<T> Display for Generated<T>
where
    T: Debug,
{
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(
            f,
            "{}: {} = {:#?}",
            self.name,
            type_name::<T>(),
            self.value.borrow().value
        )
    }
}

impl<T> GeneratedValue for Generated<T>
where
    T: 'static + Clone + Debug,
{
    fn name(&self) -> &'static str {
        self.name
    }

    fn shrink(&self, is_valid: &mut dyn FnMut() -> bool) -> bool {
        let original_value = self.value.borrow().clone();
        original_value.shrink(&mut |mut shrunk: Shrinkable<T>| {
            std::mem::swap(self.value.borrow_mut().deref_mut(), &mut shrunk);
            let valid = is_valid();
            if !valid {
                std::mem::swap(self.value.borrow_mut().deref_mut(), &mut shrunk);
            }
            valid
        })
    }

    fn value(&self) -> Box<Dynamic> {
        Dynamic::new(self.value.borrow().value.clone())
    }

    fn type_name(&self) -> &'static str {
        type_name::<T>()
    }
}

/// GeneratedValue exists to hide the real type
/// and allow for heterogenous values in the [Recording].
trait GeneratedValue: Display {
    fn name(&self) -> &'static str;
    fn type_name(&self) -> &'static str;
    fn value(&self) -> Box<Dynamic>;

    // Attempts to shrink the internal value, mutably:
    fn shrink(&self, shrink_valid: &mut dyn FnMut() -> bool) -> bool;
}

type Recording = Vec<Box<dyn GeneratedValue>>;

/// The entry-point for generating values.
pub struct Maat<'a> {
    // this type serves only to hide the Mode type,
    // which would otherwise have public members
    mode: Mode<'a>,
}

impl<'a> Maat<'a> {
    /// Generate a random value with the given `name`
    /// and using the given `generator`.
    ///
    /// The `name` is used to identify the value in
    /// the case of a failure.
    ///
    /// # Example
    /// ```rust
    /// let x = maat.generate("x", i64(0, 10_000));
    /// ```
    #[inline(always)]
    pub fn generate<T>(&mut self, name: &'static str, generator: impl Generator<T>) -> T
    where
        T: Debug + Clone + 'static,
    {
        self.mode.generate(name, generator)
    }
}

enum Mode<'a> {
    Testing {
        rng: &'a mut dyn rand::RngCore,
    },
    Recording {
        rng: &'a mut dyn rand::RngCore,
        record: &'a mut Recording,
    },
    Shrinking {
        recording_ix: usize,
        recording: &'a Recording,
    },
}

impl<'a> Mode<'a> {
    pub fn generate<T>(&mut self, name: &'static str, generator: impl Generator<T>) -> T
    where
        T: Clone + std::fmt::Debug + 'static,
    {
        match self {
            Mode::Testing { rng } => generator.generate(rng),
            Mode::Recording { rng, record } => {
                let shrinkable = generator.generate_shrinkable(rng);
                let result = shrinkable.value.clone();
                record.push(Box::new(Generated {
                    name,
                    value: RefCell::new(shrinkable),
                }));

                result
            }
            Mode::Shrinking {
                recording_ix: at,
                recording,
            } => {
                let existing = &recording[*at];
                if existing.name() != name {
                    panic!(
                        "[maat] Usage error: while shrinking, got a different name for value {at}: was {}, is {}",
                        existing.name(),
                        name
                    );
                }

                if let Some(value) = existing.value().downcast_ref::<T>() {
                    *at += 1;
                    value.clone()
                } else {
                    let old = existing.type_name();
                    let new = type_name::<T>();
                    panic!("[maat] Usage error: while shrinking, got a different type for generated value {at}: was {old}, is {new}");
                }
            }
        }
    }
}

pub struct Config {
    iterations: usize,
}

impl Default for Config {
    fn default() -> Self {
        DEFAULT_CONFIG
    }
}

const DEFAULT_CONFIG: Config = Config { iterations: 100 };

pub fn property(test: impl Fn(&mut Maat) -> bool) {
    property_cfg(test, &DEFAULT_CONFIG);
}

pub fn property_cfg(test: impl Fn(&mut Maat) -> bool, cfg: &Config) {
    // TODO: replay stored RNG values for regression-checks
    let mut rng = RNG::from_entropy();
    let start = Instant::now();
    for _ in 0..cfg.iterations {
        // store RNG state so we can reuse it for recording, if needed
        let iteration_rng = rng.clone();
        let mode = Mode::Testing { rng: &mut rng };
        if !test(&mut Maat { mode }) {
            handle_failure(test, iteration_rng);
        }
    }

    let elapsed = start.elapsed();
    println!(
        "[maat] OK, passed {} tests ({:.0} iterations/sec)",
        cfg.iterations,
        cfg.iterations as f64 / elapsed.as_secs_f64()
    );
}

#[cold]
fn handle_failure(test: impl Fn(&mut Maat) -> bool, rng: RNG) -> ! {
    let original = make_recording(&test, rng);
    let original_str = display_recording(&original);
    let shrunk = shrink_recording(&test, original);
    let shrunk_str = display_recording(&shrunk);
    panic!("\n[maat] Falsified property with values:\n{shrunk_str}\n\n[maat] Original failing values were:\n{original_str}\n");
}

fn make_recording(test: impl Fn(&mut Maat) -> bool, mut rng: RNG) -> Recording {
    let mut record = Vec::new();
    let mode = Mode::Recording {
        rng: &mut rng,
        record: &mut record,
    };

    let recording_passed = test(&mut Maat { mode });
    if recording_passed {
        panic!("[maat] Non-deterministic test function: found a failure but was unable to reproduce it.");
    }

    record
}

fn shrink_recording(test: impl Fn(&mut Maat) -> bool, recording: Recording) -> Recording {
    loop {
        let mut shrank_any = false;
        // attempt to shrink each value in the recording
        for value in &recording {
            while value.shrink(&mut || {
                // the shrink is valid if test still fails
                !test(&mut Maat {
                    mode: Mode::Shrinking {
                        recording_ix: 0,
                        recording: &recording,
                    },
                })
            }) {
                shrank_any = true;
            }
        }

        // we weren’t able to make any more shrinks
        // bail out
        if !shrank_any {
            break;
        }
    }

    recording
}

fn display_recording(recording: &Recording) -> String {
    let mut result = String::new();
    for value in recording {
        writeln!(result, "{value}").unwrap();
    }

    result
}

#[cfg(test)]
mod test {
    use super::*;
    use crate::generators::{derive, i64, string_from_example, usize};

    #[test]
    pub fn failing() {
        property(|maat| {
            let x = maat.generate("x", i64(0, 100));
            let y = maat.generate("y", i64(0, 100));
            x + y == x + x || x < 10
        })
    }

    #[test]
    pub fn add_symmetric() {
        property(|maat| {
            let x = maat.generate("x", i64(0, 10_000));
            let y = maat.generate("y", i64(0, 10_000));
            x + y == y + x
        })
    }

    #[test]
    pub fn test_string_from_example() {
        property(|maat| {
            let s = maat.generate("s", string_from_example("test", None));
            s.len() >= 3
        })
    }

    #[test]
    pub fn test_inner() {
        property(|maat| {
            let x = maat.generate("x", vec(i64(0, 100), 0, 10));
            let y = maat.generate("y", vec(i64(0, 100), 0, 10));
            x == y
        })
    }

    pub fn vec<T: 'static + Clone + std::fmt::Debug>(
        inner: impl Generator<T> + 'static,
        min_length_inclusive: usize,
        max_length_exclusive: usize,
    ) -> impl Generator<Vec<T>> {
        derive(move |maat| {
            let length = maat.generate("length", usize(min_length_inclusive, max_length_exclusive));
            let mut result = Vec::with_capacity(length);
            for _ in 0..length {
                result.push(maat.generate("element", &inner));
            }

            result
        })
    }
}
