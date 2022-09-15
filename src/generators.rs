use rand::{distributions::Distribution, Rng};

use crate::Generator;

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

        fn shrink(&self, _value: T, _shrink_valid: &mut dyn FnMut(&T) -> bool) -> Option<T> {
            None
        }
    }
}

/// The `i64` generator generates values uniformly distributed between the given bounds.
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

pub fn vec<T>(length: usize, inner: impl Generator<T>) -> impl Generator<Vec<T>> {
    struct G<Inner> {
        length: usize,
        inner: Inner,
    }

    return G { length, inner };

    impl<Inner, T> Generator<Vec<T>> for G<Inner>
    where
        Inner: Generator<T>,
    {
        fn generate(&self, rng: &mut dyn rand::RngCore) -> Vec<T> {
            let mut result = Vec::with_capacity(self.length);
            for _ in 0..self.length {
                result.push(self.inner.generate(rng));
            }

            result
        }

        fn shrink(
            &self,
            value: Vec<T>,
            shrink_valid: &mut dyn FnMut(&Vec<T>) -> bool,
        ) -> Option<Vec<T>> {
            todo!()
        }
    }
}

pub fn inner<T>(f: impl Fn(&mut crate::Maat) -> T + Clone) -> impl Generator<T> {
    struct G<T, F> {
        recording: crate::Recording,
        inner: F,
        _marker: std::marker::PhantomData<T>,
    }

    return G {
        recording: crate::Recording::new(),
        inner: f,
        _marker: std::marker::PhantomData,
    };

    impl<T, F> Generator<T> for G<T, F>
    where
        F: Fn(&mut crate::Maat) -> T,
    {
        fn generate(&self, rng: &mut dyn rand::RngCore) -> T {
            (self.inner)(&mut crate::Maat {
                mode: crate::Mode::Testing { rng },
            })
        }

        fn shrink(&self, value: T, shrink_valid: &mut dyn FnMut(&T) -> bool) -> Option<T> {
            None
        }
    }
}
