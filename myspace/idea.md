# 美化、优化JetBrains的ide工具

## # 工欲善其事必先利其器

> 一个赏心悦目的开发工具可以大大改善我们的开发心情和提升开发效率。

## # 本人的IDEA界面

![ipso](http://api.ipso.live/uploads/873e7e8ca080e6b2d6c7c112238bfd36.jpg)

## # idea为例

### 界面美化 ###

1. 美化前

![ipso](http://api.ipso.live/uploads/647974542de5cf2ad270d5d2a800afe2.jpg)

2. 美化后

![ipso](http://api.ipso.live/uploads/167f09bb12345c8f7b74a788e82cdd1b.jpg)

3. 步骤：

* 设置(Settings) -> 首选项(Appearance) -> Theme/Use custom font 设置主题/设置字体样式(非内容区)
* 设置(Settings) -> Editor -> Font 设置内容区字体样式

4. 安装新的主题插件

  本地jar包方式安装：

* 设置(Settings) -> 插件(Plugins) -> install Plugin from Disk  (这里需要自己提前下载好主题的jar包)
* 博主推荐主题包地址(JetBrains官方主题地址)：[IDEA Theme](https://plugins.jetbrains.com/contest/intellij-themes/2019)

在线下载主题插件：

* 设置(Settings) -> 插件(Plugins) -> 商城（Marketplace) 搜索主题名字点击安装(Install)

5. 博主喜欢vscode的界面风格所以博主使用的主题是(Visual Studio Code Dark Plus)

### 配置优化 ###

IDEA安装后我们可以修改部分配置文件以提高我们的开发效率，这里推荐几个配置信息

1. 自动导包：当我们调用的某个方法时可以将其所属包自动import，相反当删除该方法后其所属包也会自动从代码中删除(除该方法外在当前源文件中该包没有再被调用)

* 设置(Settings) -> Editor -> General -> Auto Import （将Add unambiguous imports on the fly和Optimize imports on the fly(for current project)两项打勾选中，然后ok就可以实现了）

2. 颜色代码高亮显示

* 设置(Settings) -> Editor -> General -> Appearance (勾选Show Css color preview as background)

![ipso](http://api.ipso.live/uploads/a9f4b1eb0f7b6b02e583d60eedea0e98.jpg)

3. 函数分割符(根据个人习惯设置)

* 设置(Settings) -> Editor -> General -> Appearance (勾选Show method separator)

4. 显示代码小地图

* 设置(Settings) -> 插件(Plugins) -> 商城（Marketplace) 搜索CodeGlance点击安装(Install)

5. 神器1--Postfix Completion(对于长期使用JetBrains IDE的同学强烈建议去理解并使用)

* 可以让你开发效率起飞的神器之一
* 这个圣器在settings的Editer的Postfix Completion中设置

6. 神器2--Live Templates(对于长期使用JetBrains IDE的同学强烈建议去理解并使用)

* 让你开发起飞的圣器之一(为什么是圣器，因为相较于Postfix Completion,它可以支持自己定义，且支持更多语言)
* 这个圣器在settings的Editer的Live Templates中设置

7. 块编辑(用过vscode的应该深有体会，功能超级强大)

8. 关联第三方框架(file -> Project Structure -> Facets), 如Spring等。

9. Local History 电脑本地，IDE提供的版本控制(控制的的是当前源文件)

* JetBrains 中默认使用的多区域选中的快捷键是：按住Ctrl+Alt+Shlft,然后鼠标左键选中需要编辑的区域

1. 最终结果

![ipso](http://api.ipso.live/uploads/553217407d935017c6db9338515497ba.jpg)

### 调试键 ###

设置断点开启dubug调试项目，断点停下后：F7转到定义；F8执行下一步；F9执行到下一个断点。

### 其它(介绍部分作者使用过的快捷键) ###

* 批量更改同一个源文件中的同一变量名或者函数名快捷键(Shlft + F6)
* 抽取变量名为函数局部变量(选中变量值Ctrl+Alt+v)
* 抽取变量名为类的成员变量(选中变量值Ctrl+Alt+f)
* 抽取函数(选中区域Ctrl+Alt+f)
* 关闭当前窗口以外的其它窗口(alt+鼠标左键)
* 从当前窗口跳转到下一个窗口(Ctrl+E)
* 快捷打开generate(alt + insert键)，包含非常高效的javaBean内容书写姿势(建议常用)
* 有继承时使用(ctrl + o),添加继承父类的方法
* 进入实现类(ctrl + alt + b)
* for遍历集合快捷操作(iter + tab)
* 接口快捷添加实现类(alt+enter)
* 快捷打出main方法(psvm)
* 快捷复制当前行到下一行(Ctrl + d)

## # 最后

通过以上设置我们不仅可以改善我们的视觉体验，还可以提高我们的编程效率。我们都知道JetBrains公司旗下有众多ide工具，如IDEA、PHPStorm、PyCharm、WebStorm、DataGrap等，其优化的方法大体类似，举一反三。
