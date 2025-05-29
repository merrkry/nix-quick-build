# nix-quick-build

Utilize the power of [nix-community/nix-eval-jobs](https://github.com/nix-community/nix-eval-jobs) to build derivations blazingly fast.

Highly inspired by [Mic92/nix-fast-build](https://github.com/Mic92/nix-fast-build/).

## Features

- [x] Retrieve eval results from `nix-eval-jobs`
- [x] Call `nix-build` to build locally
  - [ ] Correctly handle cache status in Lix
  - [ ] Distributed builds support
- [x] Push outputs to attic cache

## Side Notes

This is my first Golang project that isn't a tutorial project. lol.

It all started with a bug that stops `nix-fast-build` to actually detect cache status and push outputs.
Then I figured it might be a good idea to learn what's going on by re-inventing the wheel.
