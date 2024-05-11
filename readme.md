# cjungo

一个框架。发音“菌GO”，C 是不发音的。类似 django 的 d 不发音一样。

## [示例](https://github.com/cjungo/demo)

示例项目

## 约定俗成

使用框架给定的 Load*FromEnv 函数可以得到默认的配置，且这些配置可以通过环境变量进行配置。
再 init 方法里面使用 LoadEnv 函数可以加载 .env 文件。

配置可参考 demo 项目 example.env 文件
