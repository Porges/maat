use std::rc::Rc;

use rand::{distributions::Distribution, Rng};

use crate::{shrink_recording, Generator, Maat, Mode, Shrinkable};

/// The `placeholder` generator generates an arbitrary value that
/// doesn’t ever shrink. It is useful for generating values that are
/// needed to compile/run the test but are known not to be relevant to
/// the property being tested.
pub fn placeholder<T>() -> impl Generator<T> + Clone
where
    T: Clone,
    rand::distributions::Standard: Distribution<T>,
{
    #[derive(Clone)]
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
                    // TODO: do we need to clone the recording?
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
                            shrank_any = true
                        }
                    }
                    shrank_any
                }),
            }
        }
    }
}
