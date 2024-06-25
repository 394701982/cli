package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"gopkg.in/yaml.v2"
)

type Config struct {
	ResultsLimit int    `yaml:"results_limit"`
	Proxy        string `yaml:"proxy"`
	Concurrency  int    `yaml:"concurrency"`
}

type FofaResult struct {
	Results [][]string `json:"results"`
}

func loadConfig(configPath string) (*Config, error) {
	config := &Config{}
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func fofaSearch(apiKey, query string, resultsLimit int) ([][]string, error) {
	//	encodedQuery := base64.StdEncoding.EncodeToString([]byte(query))
	url := fmt.Sprintf("https://fofa.info/api/v1/search/all?key=%s&q=%s&size=%d", apiKey, query, resultsLimit)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result FofaResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result.Results, nil
}

func setupChromedp(proxy string) (context.Context, context.CancelFunc) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("ignore-certificate-errors", true), // 忽略证书错误
	)
	if proxy != "" {
		opts = append(opts, chromedp.ProxyServer(proxy))
	}
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel := chromedp.NewContext(allocCtx)
	return ctx, cancel
}

func getWebsiteInfo(ctx context.Context, url string) (string, int, string, string, error) {
	var buf []byte
	var pageTitle string
	var statusCode int

	// 监听网络请求事件
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if ev, ok := ev.(*network.EventResponseReceived); ok {
			if ev.Type == network.ResourceTypeDocument {
				statusCode = int(ev.Response.Status)
			}
		}
	})

	err := chromedp.Run(ctx,
		network.Enable(),
		chromedp.Navigate(url),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Title(&pageTitle),
		chromedp.CaptureScreenshot(&buf),
	)
	if err != nil {
		return url, 0, "nil", "nil", nil
		//fmt.Errorf("error navigating to %s: %v", url, err)
	}

	screenshotPath := fmt.Sprintf("screenshots/%s.png", base64.URLEncoding.EncodeToString([]byte(url)))
	if err := ioutil.WriteFile(screenshotPath, buf, 0644); err != nil {
		return "", 0, "", "", fmt.Errorf("error writing screenshot file: %v", err)
	}

	return url, statusCode, pageTitle, screenshotPath, nil
}

func main() {
	// 从命令行参数读取 FOFA API Key 和查询参数
	apiKey := flag.String("k", "", "FOFA API Key")
	query := flag.String("q", "", "Search query")
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	if *apiKey == "" || *query == "" {
		fmt.Println("API Key and query are required")
		os.Exit(1)
	}

	config, err := loadConfig(*configPath)
	if err != nil {
		fmt.Println("Error loading config file:", err)
		os.Exit(1)
	}

	results, err := fofaSearch(*apiKey, *query, config.ResultsLimit)
	if err != nil {
		fmt.Println("Error fetching FOFA results:", err)
		os.Exit(1)
	}

	err = os.MkdirAll("screenshots", 0755)
	if err != nil {
		fmt.Printf("Error creating screenshots directory: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := setupChromedp(config.Proxy)
	defer cancel()

	outputFile, err := os.Create("output.txt")
	if err != nil {
		fmt.Println("Error creating output file:", err)
		os.Exit(1)
	}
	defer outputFile.Close()

	var wg sync.WaitGroup
	sem := make(chan struct{}, config.Concurrency)

	for _, result := range results {
		url := result[0]

		wg.Add(1)
		sem <- struct{}{}

		go func(url string) {
			defer wg.Done()
			defer func() { <-sem }()

			url, statusCode, pageTitle, screenshotPath, err := getWebsiteInfo(ctx, url)
			if err != nil {
				fmt.Printf("Error processing %s: %v\n", url, err)
				return
			}

			fmt.Printf("URL: %s\nStatus Code: %d\nPage Title: %s\nScreenshot saved at: %s\n", url, statusCode, pageTitle, screenshotPath)
			outputFile.WriteString(fmt.Sprintf("%s, %d, %s, %s\n", url, statusCode, pageTitle, screenshotPath))
		}(url)
	}

	wg.Wait()
}
