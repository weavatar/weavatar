package service

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/go-rat/utils/str"
	"github.com/gookit/color"
	"github.com/urfave/cli/v3"
)

type CliService struct {
}

func NewCliService() *CliService {
	return &CliService{}
}

func (r *CliService) HashMake(ctx context.Context, cmd *cli.Command) error {
	start := uint64(10000)
	end := cmd.Uint("sum")
	dir := cmd.String("dir")
	hashType := cmd.String("type")

	color.Warnf("号最大值: %d\n", end)
	color.Warnf("存放目录: %s\n", dir)
	color.Warnf("哈希类型: %s\n\n", hashType)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	type fileInfo struct {
		file   *os.File
		writer *bufio.Writer
		mu     sync.Mutex
		count  int
	}

	files := make([]*fileInfo, 255)
	for j := uint64(0); j <= 255; j++ {
		fileName := fmt.Sprintf("%s/%d.csv", dir, j)
		file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}

		// 预分配文件空间以减少文件系统碎片
		estimatedSize := (end - start + 1) / 255 * 50 // 每行50字节
		if err = file.Truncate(int64(estimatedSize)); err != nil {
			return err
		}

		writer := bufio.NewWriterSize(file, 8*1024*1024) // 8MB缓冲区
		files[j] = &fileInfo{
			file:   file,
			writer: writer,
			count:  0,
		}
	}

	// 工作池
	numWorkers := runtime.NumCPU()
	batchSize := 10000
	workChan := make(chan struct {
		start, end uint64
	}, numWorkers)
	// 结果通道
	resultChan := make(chan struct {
		table uint64
		hash  string
		num   uint64
	}, 100000)
	// 错误通道
	errChan := make(chan error, numWorkers)
	// 完成信号
	done := make(chan struct{})

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			emailBuilder := strings.Builder{}
			emailBuilder.Grow(20) // 预分配空间

			for batch := range workChan {
				for num := batch.start; num <= batch.end; num++ {
					// 重用字符串构建器
					emailBuilder.Reset()
					emailBuilder.WriteString(strconv.FormatUint(num, 10))
					emailBuilder.WriteString("@qq.com")
					email := emailBuilder.String()

					var sum string
					if hashType == "sha256" {
						sum = str.SHA256(email)
					} else {
						sum = str.MD5(email)
					}

					table, err := strconv.ParseUint(sum[:2], 16, 64)
					if err != nil {
						errChan <- err
						return
					}

					resultChan <- struct {
						table uint64
						hash  string
						num   uint64
					}{table, sum, num}
				}
			}
		}()
	}

	// 启动写入器
	flushThreshold := 100000 // 每10万条记录刷新一次
	writeWg := sync.WaitGroup{}
	writeWg.Add(1)
	go func() {
		defer writeWg.Done()
		count := uint64(0)
		total := end - start + 1
		progressStep := total / 100 // 1%的进度
		lastProgress := uint64(0)

		for {
			select {
			case result := <-resultChan:
				file := files[result.table]
				file.mu.Lock()
				line := result.hash + "," + strconv.FormatUint(result.num, 10) + "\n"
				if _, err := file.writer.WriteString(line); err != nil {
					file.mu.Unlock()
					errChan <- err
					return
				}
				file.count++
				if file.count >= flushThreshold {
					if err := file.writer.Flush(); err != nil {
						file.mu.Unlock()
						errChan <- err
						return
					}
					file.count = 0
				}
				file.mu.Unlock()

				count++
				if count-lastProgress >= progressStep {
					color.Greenf("进度: %.2f%%\n", float64(count)/float64(total)*100)
					lastProgress = count
				}

				if count == total {
					close(done)
					return
				}

			case err := <-errChan:
				close(done)
				color.Redln("处理错误:", err)
				return
			}
		}
	}()

	// 分配工作
	go func() {
		for batch := start; batch <= end; batch += uint64(batchSize) {
			endBatch := batch + uint64(batchSize) - 1
			if endBatch > end {
				endBatch = end
			}
			workChan <- struct {
				start, end uint64
			}{batch, endBatch}
		}
		close(workChan)
	}()

	// 等待所有工作完成
	<-done
	wg.Wait()
	close(resultChan)
	writeWg.Wait()

	// 确保所有数据都刷新到磁盘
	for j := uint64(1); j <= 255; j++ {
		if err := files[j].writer.Flush(); err != nil {
			return err
		}
		if err := files[j].file.Sync(); err != nil {
			return err
		}
		if err := files[j].file.Close(); err != nil {
			return err
		}
	}

	color.Greenln("生成完成")
	return nil
}

func (r *CliService) HashInsert(ctx context.Context, cmd *cli.Command) error {
	return nil
}
