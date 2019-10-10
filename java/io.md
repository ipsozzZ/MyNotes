# IO基础

IO
IO: Input / Output

IO流是一种流式的数据输入/输出模型：

二进制数据以byte为最小单位在InputStream / OutputStream中单向流动
字符数据以char为最小单位在Reader / Writer中单向流动
JDK的java.io包提供了同步IO功能
JDK的java.nio包提供了异步IO功能
Java的IO流的接口：

字节流接口：InputStream / OutputStream
字符流接口：Reader / Writer

## # File对象

File
java.io.File表示文件系统的一个文件或者目录：

* isFile()：是否是文件
* isDirectory()：是否是目录
* 创建File对象本身不涉及IO操作

* 获取路径: getPath()
* 绝对路径: getAbsolutePath()
* 规范路径：getCanonicalPath()

**文件操作**:

* canRead()：是否允许读取该文件
* canWrite()：是否允许写入该文件
* canExecute()：是否允许运行该文件
* length()：获取文件大小
* createNewFile()：创建一个新文件
* static createTempFile()：创建一个临时文件
* delete()：删除该文件
* deleteOnExit()：在JVM退出时删除该文件

**目录操作**:

* String[] list()：列出目录下的文件和子目录名
* File[] listFiles()：列出目录下的文件和子目录名
* File[] listFiles(FileFilter filter)
* File[] listFiles(FilenameFilter filter)
* mkdir()：创建该目录
* mkdirs()：创建该目录，并在必要时将不存在的父目录也创建出来
* delete()：删除该目录

## # Input和Output

### InputStream

InputStream是所有输入流的超类：

* int read()读取一个字节
* int read(byte[])读取若干字节并填充到byte[]数组
* read()方法是阻塞（blocking）的

**注意**使用try(resource)可以保证InputStream正确关闭

**常用InputStream**:

* FileInputStream
* ByteArrayInputStream

### OutputStream

OutputStream是所有输出流的超类：

**常用方法**：

* write(int b)写入一个字节
* write(byte[])写入byte[]数组的所有字节
* flush()方法将缓冲器内容输出
* write()方法是阻塞（blocking）的

**注意**使用try(resource)可以保证OutputStream正确关闭

**常用OutputStream**：

* FileOutputStream
* ByteArrayOutputStream

### Filter模式

Filter模式是为了解决子类数量爆炸的问题。直接提供数据的InputStream：

* FileInputStream
* ByteArrayInputStream
* ServletInputStream

提供附加功能的InputStream从FilterInputStream派生：

* BufferedInputStream
* DigestInputStream
* CipherInputStream
* GZIPInputStream

Filter模式又称Decorator模式，通过少量的类实现了各种功能的组合。FilterOutputStream和FilterInputStream类似。

### 操作Zip

ZipInputStream可以读取Zip流。JarInputStream提供了额外读取jar包内容的能力。ZipOutputStream可以写入Zip流。配合FileInputStream和FileOutputStream就可以读写Zip文件。

### classpath资源

classpath中可以包含任意类型的文件。从classpath读取文件可以避免不同环境下文件路径不一致的问题。

读取classpath资源：

```java

try(InputStream input = getClass().getResourceAsStream("/default.properties")) {
  if (input != null) {
    // Read from classpath
  }
}

```

### 序列化

**序列化**是指把一个Java对象变成二进制内容（byte[]）, Java对象实现序列化必须实现Serializable接口

**反序列化**是指把一个二进制内容（byte[]）变成Java对象

使用ObjectOutputStream和ObjectInputStream实现序列化和反序列化

readObject()可能抛出的异常：

* ClassNotFoundException：没有找到对应的Class
* InvalidClassException：Class不匹配

反序列化由JVM直接构造出Java对象，不调用构造方法可设置serialVersionUID作为版本号（非必需）

## # Reader和Writer

### Reader

Reader以字符为最小单位实现了字符流输入。常用方法：

* int read() 读取下一个字符
* int read(char[]) 读取若干字符并填充到char[]数组

常用Reader类：

* FileReader：从文件读取
* CharArrayReader：从char[]数组读取

Reader是基于InputStream构造的，任何InputStream都可指定编码并通过InputStreamReader转换为Reader：```Reader reader = new InputStreamReader(input, "UTF-8")```

### Writer

Writer以字符为最小单位实现了字符流输出。常用方法：

* write(int c) 写入下一个字符
* write(char[]) 写入char[]数组的所有字符

常用Writer类：

* FileWriter：写入文件
* CharArrayWriter：写入char[]数组

Writer是基于OutputStream构造的，任何OutputStream都可指定编码并通过OutputStreamWriter转换为Writer：```Writer writer = new OutputStreamWriter(output, "UTF-8")```
