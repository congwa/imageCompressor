package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
)

func CompressImage(inputPath string, outputPath string) error {
	// 读取原始图片
	file, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return err
	}

	// 判断图片格式是否为 JPEG 或 PNG
	if format != "jpeg" && format != "png" {
		return fmt.Errorf("unsupported image format: %s", format)
	}

	// 判断是否需要转换 PNG 格式
	if format == "png" {
		// 检查是否含有透明通道
		hasAlpha, err := hasAlpha(inputPath)

		if hasAlpha {
			println("有透明度")
			// 使用 Oxipng 压缩 PNG 格式图片
			pngBytes, err := oxipngCompress(inputPath)
			if err != nil {
				return err
			}
			img, _, err = image.Decode(bytes.NewReader(pngBytes))
			if err != nil {
				return err
			}
			format = "png"
		} else {
			println("无透明度")
			// 转换 PNG 格式为 JPEG 格式
			buf := new(bytes.Buffer)
			err = imaging.Encode(buf, img, imaging.JPEG, imaging.JPEGQuality(90))

			// if err != nil {
			// 	return err
			// }
			// err = jpeg.Encode(buf, img, &jpeg.Options{Quality: 90})
			if err != nil {
				return err
			}
			img, _, err = image.Decode(buf)
			if err != nil {
				return err
			}
			format = "jpeg"
		}
	}

	// 压缩图片
	if format == "jpeg" {
		img, _ = compressJPEG(img, 90)
	}

	// 将压缩后的图片转为 WebP 格式
	webpBytes, err := webp.EncodeRGBA(img, 90)
	if err != nil {
		return err
	}

	inputBytes, err := getFileSize(inputPath)
	if err != nil {
		// 处理错误情况
		return err
	}

	// 比较压缩前后的大小并选择较小的图片
	if len(webpBytes) >= len(inputBytes) {
		// 不进行压缩，直接复制原图
		fileInfo, err := os.Stat(inputPath)
		if err != nil {
			return err
		}
		err = copyFile(inputPath, outputPath, fileInfo.Mode())
		if err != nil {
			return err
		}
	} else {
		// 使用压缩后的图片
		err = ioutil.WriteFile(outputPath, webpBytes, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func hasAlpha(filepath string) (bool, error) {
	// 打开 PNG 图片
	file, err := os.Open(filepath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// 使用 DecodeConfig 函数获取图片的详细信息
	pngConfig, err := png.DecodeConfig(file)
	if err != nil {
		return false, err
	}

	// 判断颜色模式
	if pngConfig.ColorModel == color.RGBAModel {
		return true, nil
	} else {
		return false, nil
	}
}

func getFileSize(path string) ([]byte, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	size := fileInfo.Size()
	bytes := make([]byte, size)

	return bytes, nil
}

// 使用mozJpeg压缩
func compressJPEG(img image.Image, quality int) (image.Image, error) {
	// 将图像编码为 JPEG 格式
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})

	if err != nil {
		return nil, err
	}

	// 使用 jpegtran 命令行工具来压缩 JPEG 图像
	cmd := exec.Command("jpegtran", "-optimize", "-progressive", "-copy", "none")
	cmd.Stdin = bytes.NewReader(buf.Bytes())
	var out bytes.Buffer
	cmd.Stdout = &out

	err = cmd.Run()

	if err != nil {
		return nil, err
	}

	// 解码 JPEG 图像
	compressedImg, err := jpeg.Decode(&out)

	if err != nil {
		return nil, err
	}

	return compressedImg, nil
}

func oxipngCompress(inputPath string) ([]byte, error) {
	file, err := os.Open(inputPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	buffer := new(bytes.Buffer)
	err = png.Encode(buffer, img)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("oxipng", "-o7")
	input := bytes.NewReader(buffer.Bytes())
	cmd.Stdin = input
	var output bytes.Buffer
	cmd.Stdout = &output
	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	return output.Bytes(), nil
}

// 使用 Go 语言编写的文件复制函数
func copyFile(src string, dst string, mode os.FileMode) (err error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	err = destination.Chmod(mode)
	if err != nil {
		return err
	}

	_, err = io.Copy(destination, source)
	return err
}

// func main() {
// 	inputPath := "./input/image.png"
// 	outputPath := "./output/image.webp"
// 	err := CompressImage(inputPath, outputPath)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	fmt.Println("图片压缩成功！")
// }
