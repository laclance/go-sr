# Changelog

## [1.1.13](https://github.com/laclance/go-sr/compare/v1.1.12...v1.1.13) (2026-09-08)


### Bug Fixes

* guard finite arithmetic and fuzz normalization ([#51](https://github.com/laclance/go-sr/issues/51)) ([b856069](https://github.com/laclance/go-sr/commit/b856069369055b7e98bef5b22a187ee2e5e0b9d9))

## [1.1.12](https://github.com/laclance/go-sr/compare/v1.1.11...v1.1.12) (2026-09-08)


### Performance Improvements

* reduce pathological zone clustering cost ([#45](https://github.com/laclance/go-sr/issues/45)) ([11e0082](https://github.com/laclance/go-sr/commit/11e0082c583c29b8b9d8436ac8e34d191b696b8b))

## [1.1.11](https://github.com/laclance/go-sr/compare/v1.1.10...v1.1.11) (2026-09-08)


### Bug Fixes

* handle negative prices in percentage tolerances ([#43](https://github.com/laclance/go-sr/issues/43)) ([98a7560](https://github.com/laclance/go-sr/commit/98a75602fc3300268919d60a1097bca989a93323))

## [1.1.10](https://github.com/laclance/go-sr/compare/v1.1.9...v1.1.10) (2026-09-08)


### Bug Fixes

* return exact minimum warmup history ([#41](https://github.com/laclance/go-sr/issues/41)) ([1ddb669](https://github.com/laclance/go-sr/commit/1ddb669b29d8a47a8a3b2568c670ad2b5469b5f2))

## [1.1.9](https://github.com/laclance/go-sr/compare/v1.1.8...v1.1.9) (2026-09-08)


### Bug Fixes

* guard multi-timeframe arithmetic overflow ([#39](https://github.com/laclance/go-sr/issues/39)) ([898e0a9](https://github.com/laclance/go-sr/commit/898e0a95b7213b81d067ae2f78ed2052f95f9bb1))

## [1.1.8](https://github.com/laclance/go-sr/compare/v1.1.7...v1.1.8) (2026-09-08)


### Bug Fixes

* anchor false-break reentry windows ([#37](https://github.com/laclance/go-sr/issues/37)) ([8e19dfe](https://github.com/laclance/go-sr/commit/8e19dfe53f2722be4e56ff8a41d1dce12afbb01e))

## [1.1.7](https://github.com/laclance/go-sr/compare/v1.1.6...v1.1.7) (2026-09-08)


### Bug Fixes

* anchor multi-timeframe buckets at Unix epoch ([#34](https://github.com/laclance/go-sr/issues/34)) ([29c9eab](https://github.com/laclance/go-sr/commit/29c9eab97d2e5ddd8ea061adf8213ec0a5eb05b8))

## [1.1.6](https://github.com/laclance/go-sr/compare/v1.1.5...v1.1.6) (2026-09-08)


### Bug Fixes

* make all-history sizing helpers truthful ([#32](https://github.com/laclance/go-sr/issues/32)) ([661e2c3](https://github.com/laclance/go-sr/commit/661e2c3c17b9a96b92aea46e97267fc2f1a69e06))

## [1.1.5](https://github.com/laclance/go-sr/compare/v1.1.4...v1.1.5) (2026-09-08)


### Bug Fixes

* keep zone member pivots inside geometry ([#30](https://github.com/laclance/go-sr/issues/30)) ([bf48cfe](https://github.com/laclance/go-sr/commit/bf48cfe9ca0376b09d2cbf547fb2df94f88ca8df))

## [1.1.4](https://github.com/laclance/go-sr/compare/v1.1.3...v1.1.4) (2026-09-08)


### Bug Fixes

* recheck zone dedupe replacement chains ([#25](https://github.com/laclance/go-sr/issues/25)) ([28df521](https://github.com/laclance/go-sr/commit/28df521c7c93aa505deb44460937bcb905b3df64))

## [1.1.3](https://github.com/laclance/go-sr/compare/v1.1.2...v1.1.3) (2026-09-08)


### Bug Fixes

* use warmup helper in BBGO example ([#18](https://github.com/laclance/go-sr/issues/18)) ([d0d176f](https://github.com/laclance/go-sr/commit/d0d176f0e08f94e91c3349b04de3922773fbd05f))

## [1.1.2](https://github.com/laclance/go-sr/compare/v1.1.1...v1.1.2) (2026-09-08)


### Bug Fixes

* align lookback and invalid mode semantics ([#15](https://github.com/laclance/go-sr/issues/15)) ([8541b9f](https://github.com/laclance/go-sr/commit/8541b9fe92bd9c084a5d386468f29604a3b76ffb))

## [1.1.1](https://github.com/laclance/go-sr/compare/v1.1.0...v1.1.1) (2026-09-08)


### Bug Fixes

* preserve MTF warmup across UTC alignment ([#13](https://github.com/laclance/go-sr/issues/13)) ([cfbf696](https://github.com/laclance/go-sr/commit/cfbf69645fafd56dbf281710c2efacc5f6169882))

## [1.1.0](https://github.com/laclance/go-sr/compare/v1.0.0...v1.1.0) (2026-09-02)


### Features

* relicense go-sr under Apache-2.0 ([#7](https://github.com/laclance/go-sr/issues/7)) ([d0fe4f3](https://github.com/laclance/go-sr/commit/d0fe4f338fa5f2701b707ca0d66e34f1fbffa8d8))


### Bug Fixes

* use standard Go version tags ([#8](https://github.com/laclance/go-sr/issues/8)) ([8c3e4ad](https://github.com/laclance/go-sr/commit/8c3e4ad6367a692726c1bde1b8b781fc301d1f1c))
