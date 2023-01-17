# Changelog

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
