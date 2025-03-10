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
	"gorm.io/gorm"
)

type CliService struct {
	db *gorm.DB
}

func NewCliService(db *gorm.DB) *CliService {
	return &CliService{
		db: db,
	}
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

	files := make([]*fileInfo, 256)
	for j := uint64(0); j < 256; j++ {
		fileName := fmt.Sprintf("%s/qq_%s_%d.csv", dir, hashType, j)
		file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}

		// 预分配文件空间以减少文件系统碎片
		estimatedSize := (end - start + 1) / 256 * 50 // 每行50字节
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
	for j := uint64(0); j < 256; j++ {
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
	dir := cmd.String("dir")
	hashType := cmd.String("type")

	r.db.Exec("SET GLOBAL sql_log_bin = 0")
	r.db.Exec("SET GLOBAL rocksdb_bulk_load_allow_unsorted = 1")
	r.db.Exec("SET GLOBAL rocksdb_bulk_load = 1")
	r.db.Exec("SET unique_checks = 0")
	r.db.Exec("SET GLOBAL local_infile = 1")

	for i := 0; i < 256; i++ {
		if err := r.db.Exec(fmt.Sprintf(`DROP TABLE IF EXISTS qq_%s_%d;`, hashType, i)).Error; err != nil {
			return err
		}

		color.Greenf("正在创建表: %d\n", i)
		if err := r.db.Exec(fmt.Sprintf("CREATE TABLE qq_%s_%d (h BINARY(8) NOT NULL, q BIGINT NOT NULL, PRIMARY KEY ( `h` )) ENGINE = ROCKSDB;", hashType, i)).Error; err != nil {
			return err
		}
	}

	color.Greenln("建表完成")
	color.Warnln("正在导入数据")

	for i := 0; i < 256; i++ {
		if err := r.db.Exec(fmt.Sprintf(`LOAD DATA LOCAL INFILE '%s/qq_%s_%d.csv' INTO TABLE qq_%s_%d FIELDS TERMINATED BY ',' LINES TERMINATED BY '\\n' (@h, q) SET h = UNHEX(@h);`, dir, hashType, i, hashType, i)).Error; err != nil {
			return err
		}
		color.Greenf("导入完成: qq_%s_%d\n", hashType, i)
		// 删除文件
		_ = os.Remove(fmt.Sprintf("%s/%d.csv", dir, i))
	}

	color.Warnln("导入完成")

	r.db.Exec("SET GLOBAL rocksdb_bulk_load = 0")
	r.db.Exec("SET GLOBAL rocksdb_bulk_load_allow_unsorted = 0")
	r.db.Exec("SET GLOBAL sql_log_bin = 1")
	r.db.Exec("SET unique_checks = 1")
	r.db.Exec("SET GLOBAL local_infile = 0")

	return nil
}
