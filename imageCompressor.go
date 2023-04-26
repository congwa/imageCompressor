package imageCompressor

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strconv"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
)

// 获取图片的image.Image 和图片格式
func CompressImageGetImage(img image.Image, quality int) (image.Image, string, error) {

	format, err := getImageType(img)
	if err != nil {
		return nil, "", err
	}

	// 判断图片格式是否为 JPEG 或 PNG
	if format != "jpeg" && format != "png" {
		return nil, "", fmt.Errorf("unsupported image format: %s", format)
	}

	// 判断是否需要转换 PNG 格式
	if format == "png" {
		// 检查是否含有透明通道
		_hasAlpha, err := hasAlpha(img)

		if err != nil {
			return nil, "", err
		}

		if _hasAlpha {
			println("有透明度")

			// 使用 Oxipng 压缩 PNG 格式图片
			img, _ = oxipngCompress(img)

			format = "png"
		} else {
			println("无透明度")
			// 转换 PNG 格式为 JPEG 格式
			buf := new(bytes.Buffer)
			err := imaging.Encode(buf, img, imaging.JPEG)
			if err != nil {
				return nil, "", err
			}
			img, _ = compressJPEG(buf.Bytes(), quality)
			format = "jpeg"
		}
	}
	return img, format, nil
}

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
		_hasAlpha, err := hasAlpha(img)

		if err != nil {
			return err
		}

		if _hasAlpha {
			println("有透明度")

			// 使用 Oxipng 压缩 PNG 格式图片
			img, _ = oxipngCompress(img)

			format = "png"
		} else {
			println("无透明度")
			// 转换 PNG 格式为 JPEG 格式
			buf := new(bytes.Buffer)
			// err := imaging.Encode(buf, img, imaging.JPEG)
			// if err != nil {
			// 	return err
			// }
			img, _ = compressJPEG(buf.Bytes(), 75)
		}
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
		fmt.Println("WriteFile ", outputPath)
		err = ioutil.WriteFile(outputPath, webpBytes, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func hasAlpha(img image.Image) (bool, error) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, alpha := img.At(x, y).RGBA()
			if alpha != 0xffff {
				return true, nil
			}
		}
	}
	return false, nil
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
func compressJPEG(buf []byte, quality int) (image.Image, error) {

	cmdPath := "./cjpeg_linux"

	switch runtime.GOOS {
	case "windows":
		cmdPath = "./cjpeg.exe"
	case "linux":
		cmdPath = "./cjpeg_linux"
	case "darwin":
		cmdPath = "./cjpeg_mac"
	default:
		cmdPath = "./cjpeg_linux"
	}

	args := []string{"-quality", strconv.Itoa(20)}
	cmd := exec.Command(cmdPath, args...)

	cmd.Stdin = bytes.NewReader(buf)
	var out bytes.Buffer
	cmd.Stdout = &out

	var err1 bytes.Buffer
	cmd.Stderr = &err1

	err := cmd.Run()

	if err != nil {
		fmt.Sprintf("mozjpeg failed: %s\n\n%s", err, err1.String())
		return nil, err
	}

	// 解码 JPEG 图像
	compressedImg, err := jpeg.Decode(&out)

	if err != nil {
		return nil, err
	}

	return compressedImg, nil
}

func oxipngCompress(img image.Image) (image.Image, error) {
	// 根据操作系统选择 oxipng 工具路径
	var cmdPath string
	switch runtime.GOOS {
	case "windows":
		cmdPath = "./oxipng.exe"
	case "linux":
		cmdPath = "./oxipng_linux"
	case "darwin":
		cmdPath = "./oxipng_mac"
	default:
		cmdPath = "./oxipng_linux"
	}

	fmt.Println(cmdPath)

	// 将图像编码为 PNG 格式
	buf := new(bytes.Buffer)
	// err := imaging.Encode(buf, img, imaging.PNG, imaging.PNGCompressionLevel(20))
	err := png.Encode(buf, img)
	if err != nil {
		return nil, err
	}
	// args := []string{"-o", strconv.Itoa(quality)}
	cmd := exec.Command(cmdPath)

	cmd.Stdin = bytes.NewReader(buf.Bytes())
	var out bytes.Buffer
	cmd.Stdout = &out

	var err1 bytes.Buffer
	cmd.Stderr = &err1

	err = cmd.Run()

	img, _, err = image.Decode(buf)
	return img, nil
}

func getImageType(img image.Image) (string, error) {
	switch img.(type) {
	case *image.RGBA:
		return "png", nil
	case *image.NRGBA:
		return "png", nil
	case *image.YCbCr:
		return "jpeg", nil
	case *image.CMYK:
		return "jpeg", nil
	case *image.Gray:
		return "jpeg", nil
	case *image.Gray16:
		return "png", nil
	default:
		return "", errors.New("不支持的图片类型")
	}
}

// 使用 Go 语言编写的文件复制函数
func copyFile(src string, dst string, mode os.FileMode) (err error) {
	fmt.Println("copyFile ", src)
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
