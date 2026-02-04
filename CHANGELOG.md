# Changelog

## [0.15.0](https://github.com/openkcm/api-sdk/compare/v0.14.0...v0.15.0) (2026-02-04)


### Features

* add api for event processor ([28763c6](https://github.com/openkcm/api-sdk/commit/28763c6e2b8e6b2c800686e056d1f2e6895d4428)), closes [#75](https://github.com/openkcm/api-sdk/issues/75)
* **session-manager:** GetOIDCProvider method ([#76](https://github.com/openkcm/api-sdk/issues/76)) ([6901f65](https://github.com/openkcm/api-sdk/commit/6901f65d5b8fadcb2c671adbbb0b286d91a0bb05))

## [0.14.0](https://github.com/openkcm/api-sdk/compare/v0.13.0...v0.14.0) (2025-12-18)


### Features

* updated SystemIdentifier to have type instead of region ([#70](https://github.com/openkcm/api-sdk/issues/70)) ([087d245](https://github.com/openkcm/api-sdk/commit/087d245694fd2c1dea36118e85fc97ab61fe2c1b))

## [0.13.0](https://github.com/openkcm/api-sdk/compare/v0.12.0...v0.13.0) (2025-11-26)


### Features

* **oidcmapping:** use a single jwks_uri ([#66](https://github.com/openkcm/api-sdk/issues/66)) ([3e567a9](https://github.com/openkcm/api-sdk/commit/3e567a9fd1f5153ddd0a32ed4beda8352e365038))
* validate and get sessions ([#67](https://github.com/openkcm/api-sdk/issues/67)) ([3e4c583](https://github.com/openkcm/api-sdk/commit/3e4c583df4aa5939cfab41e7792839cc7a9c9d79))


### Bug Fixes

* remove obsolete oidcprovider and fix jwks_uri ([#68](https://github.com/openkcm/api-sdk/issues/68)) ([ba19c8d](https://github.com/openkcm/api-sdk/commit/ba19c8d6cc2f90de8b71bfcf9c25eb70d142ad90))

## [0.12.0](https://github.com/openkcm/api-sdk/compare/v0.11.0...v0.12.0) (2025-11-07)


### Features

* List tenants by labels ([#60](https://github.com/openkcm/api-sdk/issues/60)) ([219b652](https://github.com/openkcm/api-sdk/commit/219b652410c75221640461927328ffff4e1084bf))

## [0.11.0](https://github.com/openkcm/api-sdk/compare/v0.10.0...v0.11.0) (2025-11-05)


### Features

* add block and unblock for auth  ([b9d39fd](https://github.com/openkcm/api-sdk/commit/b9d39fd887fdc3c2b5ade1b9ffc5b3adb0ca3416))
* add properties to ApplyOIDCMapping rpc ([#57](https://github.com/openkcm/api-sdk/issues/57)) ([1137f57](https://github.com/openkcm/api-sdk/commit/1137f57f1a79a80c7a20aca3b0abbf062d768f32))

## [0.10.0](https://github.com/openkcm/api-sdk/compare/v0.9.0...v0.10.0) (2025-10-30)


### Features

* session manager block and unblock  ([2822c62](https://github.com/openkcm/api-sdk/commit/2822c629b5c8d6cfb83dae2ddf1f2f4abd6a6859))

## [0.9.0](https://github.com/openkcm/api-sdk/compare/v0.8.1...v0.9.0) (2025-10-14)


### Features

* session manager services ([#43](https://github.com/openkcm/api-sdk/issues/43)) ([360d28e](https://github.com/openkcm/api-sdk/commit/360d28ee85d980cfb9f0a0282d9f073c5ab48851))


### Bug Fixes

* **deps:** bump google.golang.org/grpc from 1.75.1 to 1.76.0 ([#49](https://github.com/openkcm/api-sdk/issues/49)) ([cd6e1be](https://github.com/openkcm/api-sdk/commit/cd6e1bec36227434dba1244066c4a78980e461ac))
* **deps:** bump google.golang.org/protobuf from 1.36.9 to 1.36.10 ([#48](https://github.com/openkcm/api-sdk/issues/48)) ([a2b8bbc](https://github.com/openkcm/api-sdk/commit/a2b8bbcc425cec1674d2ce4f185ee3794a922c7f))

## [0.8.1](https://github.com/openkcm/api-sdk/compare/v0.8.0...v0.8.1) (2025-09-29)


### Bug Fixes

* add external id and list rpc in auth service ([0425220](https://github.com/openkcm/api-sdk/commit/04252205809074eb6b9e87920c544c51e639b4cf)), closes [#46](https://github.com/openkcm/api-sdk/issues/46)

## [0.8.0](https://github.com/openkcm/api-sdk/compare/v0.7.0...v0.8.0) (2025-09-25)


### Features

* add registry auth service ([136260d](https://github.com/openkcm/api-sdk/commit/136260de9fb62d96ae26326f09f541d20e7a9b02)), closes [#44](https://github.com/openkcm/api-sdk/issues/44)

## [0.7.0](https://github.com/openkcm/api-sdk/compare/v0.6.0...v0.7.0) (2025-09-12)


### Features

* proto add SetTenantUserGroups RPC  ([60da010](https://github.com/openkcm/api-sdk/commit/60da010dfdfecddcd0ac114d2e89d6666efc208c))

## [0.6.0](https://github.com/openkcm/api-sdk/compare/v0.5.0...v0.6.0) (2025-09-10)


### Features

* apply tenant auth ([4afb4c5](https://github.com/openkcm/api-sdk/commit/4afb4c57830f5867d780cba5b9f5b8451a5c7a56)), closes [#37](https://github.com/openkcm/api-sdk/issues/37)

## [0.5.0](https://github.com/openkcm/api-sdk/compare/v0.4.0...v0.5.0) (2025-09-05)


### Features

* add ACTION enum for tenant management operations ([#34](https://github.com/openkcm/api-sdk/issues/34)) ([15b11a8](https://github.com/openkcm/api-sdk/commit/15b11a8490e2db2c9c03545c0b09e7f41c380406))

## [0.4.0](https://github.com/openkcm/api-sdk/compare/v0.3.0...v0.4.0) (2025-08-27)


### Features

* added GetTenant RPC method ([#32](https://github.com/openkcm/api-sdk/issues/32)) ([9529ff3](https://github.com/openkcm/api-sdk/commit/9529ff34beb10e32a96aa086e05137923dfce496))

## [0.3.0](https://github.com/openkcm/api-sdk/compare/v0.2.0...v0.3.0) (2025-07-24)


### Features

* adjusts github actions ([#26](https://github.com/openkcm/api-sdk/issues/26)) ([fb384e8](https://github.com/openkcm/api-sdk/commit/fb384e8c6ebb5aa25eb02f255a17af6d520bfd06))
* move the github actions into build repo ([5d8c43f](https://github.com/openkcm/api-sdk/commit/5d8c43f7b9be8813d30d9c195751cf86b3e96334))
* Quality check validation test ([0b6552c](https://github.com/openkcm/api-sdk/commit/0b6552c9e2ee49150858664281689f952fb1af76))
* update the github action ([#20](https://github.com/openkcm/api-sdk/issues/20)) ([e1a3a20](https://github.com/openkcm/api-sdk/commit/e1a3a20b817a385a4581558420404031a08486e2))
* update the github action to sign the commits of the pull request ([#19](https://github.com/openkcm/api-sdk/issues/19)) ([95b5747](https://github.com/openkcm/api-sdk/commit/95b57476ba5bc7946ebf8dc1b9f0bd03e53d0adf))


### Bug Fixes

* **lint:** golangci change prefix ([#23](https://github.com/openkcm/api-sdk/issues/23)) ([a148482](https://github.com/openkcm/api-sdk/commit/a148482d5a74d362e64811558966a189915dc780))
