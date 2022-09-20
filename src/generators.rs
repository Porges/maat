use std::{ops::RangeBounds, rc::Rc};

use rand::{
    distributions::{DistString, Distribution},
    Rng,
};

use crate::{Generator, Maat, Mode, Shrinkable};

/// The `placeholder` generator generates an arbitrary value that
/// doesn’t ever shrink. It is useful for generating values that are
/// needed to compile/run the test but are known not to be relevant to
/// the property being tested.
pub fn placeholder<T>() -> impl Generator<T>
where
    T: Clone,
    rand::distributions::Standard: Distribution<T>,
{
    struct G<T> {
        _marker: std::marker::PhantomData<T>,
    }

    return G {
        _marker: std::marker::PhantomData,
    };

    impl<T> Generator<T> for G<T>
    where
        rand::distributions::Standard: Distribution<T>,
    {
        fn generate(&self, rng: &mut dyn rand::RngCore) -> T {
            rng.sample(rand::distributions::Standard {})
        }

        fn generate_shrinkable(&self, rng: &mut dyn rand::RngCore) -> crate::Shrinkable<T> {
            Shrinkable {
                value: self.generate(rng),
                shrink: Rc::new(|_value, _is_valid| false /* never shrinks */),
            }
        }
    }
}

macro_rules! numeric_generator {
    ($name: ident, $type:ty) => {
        /// The `$name` generator generates values uniformly distributed between the given bounds.
        pub fn $name(min_inclusive: $type, max_exclusive: $type) -> impl Generator<$type> + Copy {
            #[derive(Copy, Clone)]
            struct G {
                min_inclusive: $type,
                max_exclusive: $type,
            }

            return G {
                min_inclusive,
                max_exclusive,
            };

            impl Generator<$type> for G {
                fn generate(&self, rng: &mut dyn rand::RngCore) -> $type {
                    rng.sample(rand::distributions::Uniform::new(
                        self.min_inclusive,
                        self.max_exclusive,
                    ))
                }

                fn generate_shrinkable(&self, rng: &mut dyn rand::RngCore) -> Shrinkable<$type> {
                    let min_inclusive = self.min_inclusive;
                    Shrinkable {
                        value: self.generate(rng),
                        shrink: Rc::new(move |original_value, is_valid| {
                            let mut value = *original_value;

                            // very big shrink:
                            while value > min_inclusive && value > 0 {
                                let test = (value as f64).log10() as $type;
                                if is_valid(test) {
                                    value = test;
                                } else {
                                    break;
                                }
                            }

                            // big shrink
                            while value > min_inclusive {
                                let test = value / 2;
                                if is_valid(test) {
                                    value = test;
                                } else {
                                    break;
                                }
                            }

                            // slow shrink:
                            while value > min_inclusive {
                                let test = value - 1;
                                if is_valid(test) {
                                    value = test;
                                } else {
                                    break;
                                }
                            }

                            value != *original_value
                        }),
                    }
                }
            }
        }
    };
}

numeric_generator!(i64, i64);
numeric_generator!(i32, i32);
numeric_generator!(i16, i16);
numeric_generator!(i8, i8);
numeric_generator!(u64, u64);
numeric_generator!(u32, u32);
numeric_generator!(u16, u16);
numeric_generator!(u8, u8);
numeric_generator!(usize, usize);
numeric_generator!(isize, isize);

pub fn string<B: RangeBounds<usize>>(bounds: B) -> impl Generator<String> {
    struct G {
        min_len_inclusive: usize,
        max_len_inclusive: usize,
    }

    let min_len_inclusive = match bounds.start_bound() {
        std::ops::Bound::Included(&x) => x,
        std::ops::Bound::Excluded(&x) => x + 1,
        std::ops::Bound::Unbounded => 0,
    };

    let max_len_inclusive = match bounds.end_bound() {
        std::ops::Bound::Included(&x) => x,
        std::ops::Bound::Excluded(&x) => x - 1,
        std::ops::Bound::Unbounded => usize::MAX, // uhhhhhhhhhhhh
    };

    return G {
        min_len_inclusive,
        max_len_inclusive,
    };

    impl Generator<String> for G {
        fn generate(&self, rng: &mut dyn rand::RngCore) -> String {
            let length = rng.sample(rand::distributions::Uniform::new_inclusive(
                self.min_len_inclusive,
                self.max_len_inclusive,
            ));

            rand::distributions::Standard {}.sample_string(rng, length)
        }

        fn generate_shrinkable(&self, rng: &mut dyn rand::RngCore) -> Shrinkable<String> {
            Shrinkable {
                value: self.generate(rng),
                shrink: Rc::new(|original_value, is_valid| todo!()),
            }
        }
    }
}

pub fn string_from_example(
    value: &str,
    max_length_inclusive: Option<usize>,
) -> impl Generator<String> {
    struct G {
        input: Vec<u8>,
        max_length_inclusive: Option<usize>,
    }

    return G {
        input: value.as_bytes().to_owned(),
        max_length_inclusive,
    };

    impl Generator<String> for G {
        fn generate(&self, rng: &mut dyn rand::RngCore) -> String {
            let seed = Some(rng.next_u32());
            let bytes = radamsa::generate(&self.input, seed, self.max_length_inclusive);
            let result = String::from_utf8_lossy(&bytes).to_string();
            result
        }

        fn generate_shrinkable(&self, rng: &mut dyn rand::RngCore) -> Shrinkable<String> {
            Shrinkable {
                value: self.generate(rng),
                // TODO: use string shrinking
                shrink: Rc::new(|_original_value, _is_valid| false),
            }
        }
    }
}

pub fn string_alphanumeric(length: usize) -> impl Generator<String> {
    struct G {
        length: usize,
    }

    return G { length };

    impl Generator<String> for G {
        fn generate(&self, rng: &mut dyn rand::RngCore) -> String {
            rand::distributions::Standard {}.sample_string(rng, self.length)
        }

        fn generate_shrinkable(&self, rng: &mut dyn rand::RngCore) -> Shrinkable<String> {
            Shrinkable {
                value: self.generate(rng),
                shrink: Rc::new(|original_value, is_valid| todo!()),
            }
        }
    }
}

pub fn alphanumeric() -> impl Generator<char> {
    struct G {}

    return G {};

    impl Generator<char> for G {
        fn generate(&self, rng: &mut dyn rand::RngCore) -> char {
            rng.sample(rand::distributions::Alphanumeric {}) as char
        }

        fn generate_shrinkable(&self, rng: &mut dyn rand::RngCore) -> Shrinkable<char> {
            Shrinkable {
                value: self.generate(rng),
                shrink: Rc::new(|original_value, is_valid| {
                    let mut value = *original_value;

                    // generator only generates u8s so it is safe to cast them

                    // shrink towards X
                    while value > 'x' {
                        let new_value = ((value as u8) - 1) as char;
                        if is_valid(new_value) {
                            value = new_value;
                        } else {
                            break;
                        }
                    }

                    while value < 'x' {
                        let new_value = ((value as u8) + 1) as char;
                        if is_valid(new_value) {
                            value = new_value;
                        } else {
                            break;
                        }
                    }

                    value != *original_value
                }),
            }
        }
    }
}

pub fn derive<T>(f: impl Fn(&mut crate::Maat) -> T + 'static) -> impl Generator<T> {
    struct G<T, F> {
        deriver: Rc<F>,
        _marker: std::marker::PhantomData<T>,
    }

    return G {
        deriver: Rc::new(f),
        _marker: std::marker::PhantomData,
    };

    impl<T, F> Generator<T> for G<T, F>
    where
        F: Fn(&mut Maat) -> T + 'static,
    {
        fn generate(&self, rng: &mut dyn rand::RngCore) -> T {
            let mode = Mode::Testing { rng };
            (self.deriver)(&mut Maat { mode })
        }

        fn generate_shrinkable(&self, rng: &mut dyn rand::RngCore) -> Shrinkable<T> {
            let mut recording = Vec::new();
            let mode = Mode::Recording {
                rng,
                record: &mut recording,
            };

            let deriver = self.deriver.clone();
            let value = deriver(&mut crate::Maat { mode });
            Shrinkable {
                value,
                shrink: Rc::new(move |_original_value, is_valid| {
                    let mut ever_shrank = false;
                    loop {
                        let mut shrank_any = false;
                        for v in &recording {
                            while v.shrink(&mut || {
                                is_valid(deriver(&mut Maat {
                                    mode: Mode::Shrinking {
                                        recording_ix: 0,
                                        recording: &recording,
                                    },
                                }))
                            }) {
                                shrank_any = true;
                            }
                        }

                        if !shrank_any {
                            break;
                        } else {
                            ever_shrank = true;
                        }
                    }

                    ever_shrank
                }),
            }
        }
    }
}
