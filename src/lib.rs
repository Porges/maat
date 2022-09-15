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
//! [maat] Shrunk failure:
//! x: i64 = 10
//! y: i64 = 0
//!
//! [maat] Original failure:
//! x: i64 = 49
//! y: i64 = 27
//!
//! ', src/lib.rs:287:13
//! ```

use dynamic::Dynamic;
use rand::SeedableRng;
use rand_xoshiro::Xoshiro256PlusPlus as RNG;
use std::{
    any::type_name,
    cell::RefCell,
    fmt::{Debug, Display, Write},
    time::Instant,
};

pub mod generators;

pub struct Shrinkable<T> {
    value: T,
    shrink: Box<dyn Fn(&mut dyn FnMut(&T) -> bool) -> Option<T>>,
}

pub trait Generator<T> {
    // Fast path, used during Testing:
    fn generate(&self, rng: &mut dyn rand::RngCore) -> T;

    // Slower path, used during Recording:
    //fn generate_shrinkable(&self, rng: &mut dyn rand::RngCore) -> Box<Shrinkable<T>>;

    fn shrink(&self, value: T, shrink_valid: &mut dyn FnMut(&T) -> bool) -> Option<T>;
}

#[derive(Clone)]
struct Generated<T, G> {
    name: &'static str,
    value: RefCell<dynamic::Described<T>>,
    generator: G,
}

impl<T, G> Display for Generated<T, G>
where
    T: Debug,
{
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(
            f,
            "{}: {} = {:#?}",
            self.name,
            type_name::<T>(),
            self.value.borrow().data
        )
    }
}

impl<T, G> GeneratedValue for Generated<T, G>
where
    T: 'static + Clone + Debug,
    G: Generator<T> + 'static + Clone,
{
    fn name(&self) -> &'static str {
        self.name
    }

    fn shrink(&self, is_valid: &dyn Fn() -> bool) -> bool {
        let old_self = self.value.borrow().data.clone();
        if let Some(value) = {
            let initial_value = self.value.borrow().data.clone();
            self.generator.shrink(initial_value, &mut |shrunk: &T| {
                self.value.borrow_mut().data = shrunk.clone();
                is_valid()
            })
        } {
            /*
            println!(
                "shrank {} from {} to {}",
                self.name,
                self.value.borrow().data,
                value
            );
            */
            self.value.borrow_mut().data = value;
            true
        } else {
            self.value.borrow_mut().data = old_self;
            false
        }
    }

    fn value(&self) -> Box<Dynamic> {
        Dynamic::new(self.value.borrow().data.clone())
    }

    fn type_name(&self) -> &'static str {
        type_name::<T>()
    }
}

pub trait GeneratedValue: Display {
    fn name(&self) -> &'static str;
    fn type_name(&self) -> &'static str;
    fn value(&self) -> Box<Dynamic>;

    // Note that this uses interior mutation, sorry:
    fn shrink(&self, shrink_valid: &dyn Fn() -> bool) -> bool;
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
    pub fn generate<T>(
        &mut self,
        name: &'static str,
        generator: impl Generator<T> + 'static + Clone,
    ) -> T
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
    pub fn generate<T>(
        &mut self,
        name: &'static str,
        generator: impl Generator<T> + 'static + Clone,
    ) -> T
    where
        T: Clone + std::fmt::Debug + 'static,
    {
        match self {
            Mode::Testing { rng } => generator.generate(rng),
            Mode::Recording { rng, record } => {
                let value = generator.generate(rng);
                record.push(Box::new(Generated {
                    name,
                    value: RefCell::new(dynamic::Described::new(value.clone())),
                    generator,
                }));
                value
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
        "[maat] OK, passed {} tests ({} iterations/s)",
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
    panic!("\n[maat] Shrunk failure:\n{shrunk_str}\n\n[maat] Original failure:\n{original_str}\n");
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
        for ix in 0..recording.len() {
            let gv = &recording[ix];
            while gv.shrink(&|| {
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
    use crate::generators::i64;

    #[test]
    pub fn failing() {
        property(|maat| {
            let x = maat.generate("x", i64(0, 100));
            let y = maat.generate("y", i64(0, 100));
            x + y == x + x || x < 10
        });
    }

    #[test]
    pub fn add_symmetric() {
        property_cfg(
            |maat| {
                let x = maat.generate("x", i64(0, 10_000));
                let y = maat.generate("y", i64(0, 10_000));
                x + y == y + x
            },
            &Config {
                iterations: 37_000_000,
            },
        );
    }
}
