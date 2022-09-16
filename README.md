# 𓁦 maat 𓆄 

maat is an experiment

----

Write a test that fails:

```rust
#[test]
pub fn test_inner() {
    property(|maat| {
        let x = maat.generate("x", vec(i64(0, 100), 0, 10));
        let mut y = x.clone();
        y.reverse();
        x == y
    })
}
```

Run it:

```
[maat] Falsified property with values:
x: alloc::vec::Vec<i64> = [
    0,
    1,
]


[maat] Original failing values were:
x: alloc::vec::Vec<i64> = [
    51,
    32,
    90,
    50,
    33,
    97,
    61,
    77,
]
```

Thanks `maat`.
