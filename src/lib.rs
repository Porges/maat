use dynamic::Dynamic;
use rand::{prelude::Distribution, Rng, SeedableRng};
use rand_xoshiro::Xoshiro256PlusPlus as RNG;
use std::{
    any::type_name,
    cell::RefCell,
    fmt::{Display, Write},
};

const MAAT: &str = "𓁦";

pub trait Generator<T> {
    fn generate(&self, rng: &mut dyn rand::RngCore) -> T;
    fn shrink(&self, value: T, shrink_valid: &mut dyn FnMut(&T) -> bool) -> Option<T>;
}

pub fn i64(min_inclusive: i64, max_exclusive: i64) -> impl Generator<i64> + Clone {
    #[derive(Clone)]
    struct G {
        min_inclusive: i64,
        max_exclusive: i64,
    }

    return G {
        min_inclusive,
        max_exclusive,
    };

    impl Generator<i64> for G {
        fn generate(&self, rng: &mut dyn rand::RngCore) -> i64 {
            rng.sample(rand::distributions::Uniform::new(
                self.min_inclusive,
                self.max_exclusive,
            ))
        }

        fn shrink(&self, value: i64, shrink_valid: &mut dyn FnMut(&i64) -> bool) -> Option<i64> {
            let mut v = value;

            // very big shrink:
            while v > self.min_inclusive && v > 0 {
                let test = (v as f64).log10() as i64;
                if shrink_valid(&test) {
                    v = test;
                } else {
                    break;
                }
            }

            // big shrink
            while v > self.min_inclusive {
                let test = v / 2;
                if shrink_valid(&test) {
                    v = test;
                } else {
                    break;
                }
            }

            // slow shrink:
            while v > self.min_inclusive {
                let test = v - 1;
                if shrink_valid(&test) {
                    v = test;
                } else {
                    break;
                }
            }

            if v != value {
                Some(v)
            } else {
                None
            }
        }
    }
}

#[derive(Clone)]
struct Generated<T, G> {
    name: &'static str,
    value: RefCell<dynamic::Described<T>>,
    generator: G,
}

impl<T, G> GeneratedValue for Generated<T, G>
where
    T: 'static + Clone + Display,
    G: Generator<T> + 'static + Clone,
{
    fn name(&self) -> &'static str {
        self.name
    }

    fn shrink(&self, is_valid: &dyn Fn() -> bool) -> bool {
        if let Some(value) = {
            let initial_value = self.value.borrow().data.clone();
            self.generator.shrink(initial_value, &mut |shrunk: &T| {
                let old_self = self.value.borrow().data.clone();
                self.value.borrow_mut().data = shrunk.clone();
                let passed = is_valid();
                self.value.borrow_mut().data = old_self;
                passed
            })
        } {
            println!(
                "shrank {} from {} to {}",
                self.name,
                self.value.borrow().data,
                value
            );
            self.value.borrow_mut().data = value;
            true
        } else {
            false
        }
    }

    fn value(&self) -> Box<Dynamic> {
        Dynamic::new(self.value.borrow().data.clone())
    }

    fn set_value(&mut self, value: Box<Dynamic>) {
        self.value.borrow_mut().data = value.downcast().expect("value of wrong type provided").data;
    }

    fn cloned(&self) -> Box<dyn GeneratedValue> {
        Box::new(self.clone())
    }

    fn display(&self) -> String {
        self.value.borrow().data.to_string()
    }

    fn type_name(&self) -> &'static str {
        type_name::<T>()
    }
}

pub trait GeneratedValue {
    fn name(&self) -> &'static str;
    fn type_name(&self) -> &'static str;
    fn shrink(&self, shrink_valid: &dyn Fn() -> bool) -> bool;
    fn display(&self) -> String;
    fn value(&self) -> Box<Dynamic>;
    fn set_value(&mut self, value: Box<Dynamic>);
    fn cloned(&self) -> Box<dyn GeneratedValue>;
}

type Recording = Vec<Box<dyn GeneratedValue>>;

pub enum Maat<'a> {
    Testing {
        rng: &'a mut dyn rand::RngCore,
    },
    Recording {
        rng: &'a mut dyn rand::RngCore,
        record: &'a mut Recording,
    },
    Shrinking {
        at: usize,
        recording: &'a Recording,
    },
}

impl<'a> Maat<'a> {
    pub fn generate<T>(
        &mut self,
        name: &'static str,
        generator: impl Generator<T> + 'static + Clone,
    ) -> T
    where
        T: Clone + std::fmt::Display + 'static,
    {
        match self {
            Maat::Testing { rng } => generator.generate(rng),
            Maat::Recording { rng, record } => {
                let value = generator.generate(rng);
                record.push(Box::new(Generated {
                    name,
                    value: RefCell::new(dynamic::Described::new(value.clone())),
                    generator,
                }));
                value
            }
            Maat::Shrinking { at, recording } => {
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

pub trait Context {
    fn arbitrary<T>(&self, name: &'static str) -> T
    where
        rand::distributions::Standard: Distribution<T>,
    {
        self.generate(name, rand::distributions::Standard {})
    }

    fn generate<T>(&self, name: &'static str, distr: impl Distribution<T>) -> T;
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
    for _ in 0..cfg.iterations {
        // store RNG state so we can reuse it for recording, if needed
        let record_rng = rng.clone();
        let passed = test(&mut Maat::Testing { rng: &mut rng });
        if !passed {
            // we got a test failure, make a recording:
            let original_recording = make_recording(&test, record_rng);

            // now shrink the recording
            let shrunk_recording = shrink_recording(
                &test,
                original_recording.iter().map(|x| x.cloned()).collect(),
            );

            let mut original = String::new();
            for value in original_recording {
                writeln!(original, "{}: {}", value.name(), value.display()).unwrap();
            }

            let mut shrunk = String::new();
            for value in shrunk_recording {
                writeln!(shrunk, "{}: {}", value.name(), value.display()).unwrap();
            }

            panic!("\n[maat] Shrunk failure:\n{shrunk}\n\n[maat] Original failure:\n{original}\n");
        }
    }

    println!("{MAAT} OK, passed {} tests", cfg.iterations);
}

fn make_recording(test: impl Fn(&mut Maat) -> bool, mut rng: RNG) -> Recording {
    let mut record = Vec::new();
    let mut recording = Maat::Recording {
        rng: &mut rng,
        record: &mut record,
    };

    let recording_passed = test(&mut recording);
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
                !test(&mut Maat::Shrinking {
                    at: 0,
                    recording: &recording,
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

#[cfg(test)]
mod test {
    use super::*;

    #[test]
    pub fn run() {
        property(|maat| {
            let x = maat.generate("x", i64(0, 100));
            let y = maat.generate("y", i64(0, 100));
            x + y == x + x || x < 10
        });
    }
}
