package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/linchengzhi/lottery/domain/dto"
	"github.com/linchengzhi/lottery/util"
	"io"
	"net/http"
	"sync"
	"time"
)

type Result struct {
	Success int
	Failure int
}

func main() {
	duration := 60 * time.Second
	requestsPerSecond := 500

	result := runLoadTest(duration, requestsPerSecond)

	fmt.Printf("Load test completed.\nSuccessful requests: %d\nFailed requests: %d\n", result.Success, result.Failure)
}

func runLoadTest(duration time.Duration, requestsPerSecond int) Result {
	result := Result{}
	var wg sync.WaitGroup
	resultChan := make(chan bool)

	go collectResults(&result, resultChan)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	endTime := time.Now().Add(duration)

	var userId int64 = 1
	for time.Now().Before(endTime) {
		<-ticker.C
		for i := 0; i < requestsPerSecond; i++ {
			userId += 1
			wg.Add(1)
			go func(userId int64) {
				defer wg.Done()
				success := sendRequest(userId)
				resultChan <- success
			}(userId)
		}
	}

	wg.Wait()
	close(resultChan)

	return result
}

func sendRequest(userId int64) bool {
	url := "http://127.0.0.1:8080/lottery/draw"

	requestId := generateRequestId()

	// 创建一个新的 DrawReq 实例
	payload := dto.DrawReq{
		RequestTime: time.Now(),
		UserId:      userId,
		ActivityId:  12345, // 假设活动ID是固定的，如果需要可以修改
		DrawNum:     10,    // 假设每次抽奖次数为1，如果需要可以修改
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return false
	}

	// 创建一个新的请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return false
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("request_id", util.UUID())

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return false
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return false
	}

	// 打印响应状态码和内容
	fmt.Printf("Request ID: %s\n", requestId)
	fmt.Printf("Response Status: %s\n", resp.Status)
	fmt.Printf("Response Body: %s\n", string(body))

	return resp.StatusCode == http.StatusOK
}

func generateRequestId() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func collectResults(result *Result, resultChan <-chan bool) {
	for success := range resultChan {
		if success {
			result.Success++
		} else {
			result.Failure++
		}
	}
}
