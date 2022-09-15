use dynamic::Dynamic;
use rand::{prelude::Distribution, Rng, SeedableRng};
use rand_xoshiro::Xoshiro256PlusPlus as RNG;
use std::fmt::{Display, Write};

const MAAT: &str = "𓁦";

pub trait Generator<T> {
    fn generate(&self, rng: &mut dyn rand::RngCore) -> T;
    fn shrink(&self, value: &T) -> Option<T>;
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

        fn shrink(&self, value: &i64) -> Option<i64> {
            let v = *value;
            if v > self.min_inclusive {
                if v > 0 {
                    return Some((v as f64).log10() as i64);
                }

                let r = v / 2;
                if r >= self.min_inclusive {
                    return Some(r);
                }

                let r = v - 1;
                return Some(r);
            }

            None
        }
    }
}

#[derive(Clone)]
struct Generated<T, G> {
    name: &'static str,
    value: dynamic::Described<T>,
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

    fn shrink(&mut self) -> bool {
        if let Some(value) = self.generator.shrink(&self.value.data) {
            println!("shrank {} from {} to {}", self.name, self.value.data, value);
            self.value.data = value;
            true
        } else {
            false
        }
    }

    fn value(&self) -> Box<Dynamic> {
        Dynamic::new(self.value.data.clone())
    }

    fn set_value(&mut self, value: Box<Dynamic>) {
        self.value.data = value.downcast().expect("value of wrong type provided").data;
    }

    fn cloned(&self) -> Box<dyn GeneratedValue> {
        Box::new(self.clone())
    }

    fn display(&self) -> String {
        self.value.data.to_string()
    }
}

pub trait GeneratedValue {
    fn name(&self) -> &'static str;
    fn shrink(&mut self) -> bool;
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
                    value: dynamic::Described::new(value.clone()),
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
                    let old = existing.value().id();
                    let new = core::any::TypeId::of::<T>();
                    panic!("[maat] Usage error: while shrinking, got a different type for generated value {at}: was {old:?}, is {new:?}");
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

fn shrink_recording(test: impl Fn(&mut Maat) -> bool, mut recording: Recording) -> Recording {
    loop {
        let mut shrank_any = false;
        for ix in 0..recording.len() {
            let mut old_value: Box<Dynamic> = recording[ix].value();
            while recording[ix].shrink() {
                // TODO: we need to be able to try a series of shrinks
                // at the moment we only try the biggest shrink each time

                // this is worse than the go version

                // the shrinker really needs to be able to return a list of shrinks
                // (biggest shrink first)

                let passed = test(&mut Maat::Shrinking {
                    at: 0,
                    recording: &mut recording,
                });

                if passed {
                    // we shrank too far
                    println!("shrank too far");
                    recording[ix].set_value(old_value);
                    break;
                }

                shrank_any = true;
                old_value = recording[ix].value();
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
