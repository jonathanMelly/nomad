# Changelog

## [1.2.0](https://github.com/jonathanMelly/nomad/compare/v1.1.1...v1.2.0) (2023-03-11)


### Features

* **apps:** new apps (still to be validated/completed) ([705be0e](https://github.com/jonathanMelly/nomad/commit/705be0e54b377a5fb6484300ed1b39c1df0b4a93))
* **cmd:** add install/update/status for further improvements ([572ef61](https://github.com/jonathanMelly/nomad/commit/572ef614f4cb18068ef60d886e8656010fc1f49b))
* **config:** added basic config to handle GitHub token for API ([c32d714](https://github.com/jonathanMelly/nomad/commit/c32d7149670803a1b10c3cbe517610a83ce2f7a6))
* **definitions:** 3 levels of appdefs (global conf, custom,embedded). ([39cbd5f](https://github.com/jonathanMelly/nomad/commit/39cbd5fac12bffdd948cb4bd6334ae5f08a9729b))
* **go:** upgrade to 1.20 and removed useless local mod rename... ([823c627](https://github.com/jonathanMelly/nomad/commit/823c627c4c142fcffdb6f423ee663a8ab147a6b6))
* **latest:** useLatest is now a cmd flag (still available on config) ([bc3fe42](https://github.com/jonathanMelly/nomad/commit/bc3fe427a59756527fbf6d8f0b22ae1ce70a6df6))
* **putty:** only 1 instance at a time ([5083937](https://github.com/jonathanMelly/nomad/commit/5083937596699d7d895d5fe275a7fc0bc452a62e))
* **readme:** pas -&gt; nomad ([a289fe9](https://github.com/jonathanMelly/nomad/commit/a289fe9055a8a52660962b7583da805cd88e5a9a))
* **readme:** revamped readme with the latest news ([6f9aaa6](https://github.com/jonathanMelly/nomad/commit/6f9aaa69e45714770f440e064b53b573f0ce501e))
* **standalone:** first alpha of a standalone installer (should run out of the box) ([bdc58b5](https://github.com/jonathanMelly/nomad/commit/bdc58b50f91470a5fe1d8c0b068f17daae094b4e)), closes [#5](https://github.com/jonathanMelly/nomad/issues/5)
* **symlink:** use junction in windows (available to any user without any conf..) ([30f3227](https://github.com/jonathanMelly/nomad/commit/30f3227f36b01577b0467c3f3c66c40f552e6944))
* **tests:** added automated tests basics ([572ef61](https://github.com/jonathanMelly/nomad/commit/572ef614f4cb18068ef60d886e8656010fc1f49b))
* **version:** complete version parsing for action decision (upgrade, downgrade) ([a9c8f7b](https://github.com/jonathanMelly/nomad/commit/a9c8f7b20b07ff4fbb1e8b85205e619cdffe8728))
* **version:** full version parts parsing (and available for app config) ([572ef61](https://github.com/jonathanMelly/nomad/commit/572ef614f4cb18068ef60d886e8656010fc1f49b))


### Bug Fixes

* **ccleaner:** moved to full toml (discard json) ([3b648c0](https://github.com/jonathanMelly/nomad/commit/3b648c0d96a45bcf60f5540235f38dd57b457284))
* **devmode:** add devmode info on readme ([6969967](https://github.com/jonathanMelly/nomad/commit/6969967cb907e99304dc7b28da6a446e4203c993))
* **download:** no more error on download error ([ff24f2d](https://github.com/jonathanMelly/nomad/commit/ff24f2d79a3d19493984693ca136b3a21fafac7e))
* **ghapikey:** get from correct config ([f09b8be](https://github.com/jonathanMelly/nomad/commit/f09b8becae6e31895d12908335877deaeb4641ed))

## [1.1.1](https://github.com/jonathanMelly/portable-app-installer/compare/v1.1.0...v1.1.1) (2023-01-20)


### Bug Fixes

* **release:** archive not anymore empty ;-) ([e46e4e4](https://github.com/jonathanMelly/portable-app-installer/commit/e46e4e41a5de50e2e72f8a75346a2e3e93a7270e))

## [1.1.0](https://github.com/jonathanMelly/portable-app-installer/compare/v1.0.0...v1.1.0) (2023-01-17)


### Features

* **apps:** add app defs in release ([eecdfa7](https://github.com/jonathanMelly/portable-app-installer/commit/eecdfa7532fe907eabb8674363eb4e985807ed4b))
* **apps:** bunch of new apps ([0780590](https://github.com/jonathanMelly/portable-app-installer/commit/07805909c226c70b57c7c1c69f518285e9428afb))
* **archive format:** nearly ready for other formats... ([3d06c49](https://github.com/jonathanMelly/portable-app-installer/commit/3d06c49355944f6996ad8c4283c8fa3f3973f630))
* **putty:** use putty cac (handles security keys) ([6cf13bc](https://github.com/jonathanMelly/portable-app-installer/commit/6cf13bce645ee70af96f05ea84552021f1a1deb1))
* **release:** add configs to zip ([639725d](https://github.com/jonathanMelly/portable-app-installer/commit/639725d06f19205f99e983396be466ae5be518b6))
* **version:** versions parts available to json configs ([4fc59f6](https://github.com/jonathanMelly/portable-app-installer/commit/4fc59f6a96f14c040cf626487ff9e8399912b5e9))


### Bug Fixes

* **download:** correctly replace version number in url ([bdc1497](https://github.com/jonathanMelly/portable-app-installer/commit/bdc149719a104e361a20fc862dc32793e5d691c0))
* **vlc:** working portable config ([f39ef45](https://github.com/jonathanMelly/portable-app-installer/commit/f39ef45ff4c9ce42793b06a73b7e7a96571c06d7))

## 1.0.0 (2023-01-09)


### Features

* **cmd:** add forecExtract/skipDownload to cmd options ([3b185ba](https://github.com/jonathanMelly/portable-app-installer/commit/3b185ba273c316c50a8d22f46fa3fa69443559b0))
* **extract:** add failing regexp on error ([e46ee95](https://github.com/jonathanMelly/portable-app-installer/commit/e46ee9516accdc0d0aca9575e4f5260fe8ea28f3))
* **git:** added mingit-busybox ([0f6ba6a](https://github.com/jonathanMelly/portable-app-installer/commit/0f6ba6a8ef2885f8f063f9a14b2eff84027f7f08))
* **latest version:** add a {{VERSIon}} placeholder for default version regex ([0f6ba6a](https://github.com/jonathanMelly/portable-app-installer/commit/0f6ba6a8ef2885f8f063f9a14b2eff84027f7f08))
* **restore:** restore any file from previous version (usefull for configs) ([3b185ba](https://github.com/jonathanMelly/portable-app-installer/commit/3b185ba273c316c50a8d22f46fa3fa69443559b0))
* **shortcut:** generate windows shortcut with optional envvar ([3b185ba](https://github.com/jonathanMelly/portable-app-installer/commit/3b185ba273c316c50a8d22f46fa3fa69443559b0))
* **symlink:** create symlink for latest version ([3b185ba](https://github.com/jonathanMelly/portable-app-installer/commit/3b185ba273c316c50a8d22f46fa3fa69443559b0))
* **version:** handles versions numbers without dot, ex 606 for ccleaner ([cad83e7](https://github.com/jonathanMelly/portable-app-installer/commit/cad83e7faca0720e6080b10d4b8cddc62df93199))


### Bug Fixes

* **download-403:** add http headers to avoid 403 ([54599d6](https://github.com/jonathanMelly/portable-app-installer/commit/54599d6659634860f1f587b3de4c1f98b1c3a23f))
* **putty:** do not delete config ([5037316](https://github.com/jonathanMelly/portable-app-installer/commit/5037316115748f539fd88b4afb15bb949a88ad23))
* **restore:** avoid trying to restore directory as  a file ([3d5d9f2](https://github.com/jonathanMelly/portable-app-installer/commit/3d5d9f2cd02f9ff7a80fc69f67bbf730407d3456))
* **shortcut:** ok if no custom icon ([a1d2070](https://github.com/jonathanMelly/portable-app-installer/commit/a1d20701560497f9cd1a296af67750833b599e2a))
