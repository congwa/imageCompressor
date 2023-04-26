# imageCompressor
lossless compress images and convert to WebP format

一个简单的包，专门用于**无损**压缩图片并转换为 WebP 格式。

## 压缩效果

![Alt text](/imgs/result.png)
**5.2MB -> 223KB**

## 使用

```golang
//导入包
got get -u https://github.com/congwa/imageCompressor

package imageCompressor

// 使用1
// 参数 image.Image
img,format,err := imageCompressor.CompressImageGetImage(image image.Image)

// img为图片的image.Image对象， format为图片的格式 png或者jpeg

// 使用2
// 传入要压缩的图片路径 和要放入的图片路径  需要绝对路径
err := imageCompressor.CompressImageGetImage(inputPath, outPutPath)


```

## 功能

- 压缩图片以减小文件大小
- 将图片转换为高效的 [WebP 格式](https://developers.google.com/speed/webp)，加快加载速度
- 支持 `.jpg` 和 `.png` 两种图片格式

## 注意事项

- 图片必须以文件路径的形式提供。
- 输出文件的扩展名始终为 `.webp`。

## 其它语言版本

[node版本](https://github.com/congwa/image-lossless-compressor)


## 如果你想了解更多关于图片处理的事情

[图片处理以wasm形式在浏览器处理](https://github.com/congwa/wasm-codecs-browser)

同时，欢迎来我的[博客](https://github.com/congwa/Front-end-Basics-Notes)转转，你会和我一起成长。



## 许可证

[MIT](https://opensource.org/licenses/MIT)