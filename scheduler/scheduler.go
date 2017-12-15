package scheduler

import (
	"sync"
	"github.com/tokwii/crawl/queue"
	"github.com/tokwii/crawl/fetcher"
	"github.com/tokwii/crawl/storage"
	"fmt"
)

type Scheduler struct{
	CStorage *storage.CrawlerStorage
	CQueue *queue.CrawlerQueue
	numWorkers int
}

func InitSchedule(numWorkers int, seedUrls []string)(*Scheduler){
	var s Scheduler
	s.CQueue = queue.InitCrawlerQueue()
	s.CStorage = storage.InitCrawlerStorage()
	s.numWorkers = numWorkers

	for _, url := range seedUrls{
		s.CQueue.Push(url)
	}

	return &s
}

func (s *Scheduler) Schedule(){
	s.initCrawlWorkerPool()
	s.CQueue.Close()
}

func (s *Scheduler) initCrawlWorkerPool(){
	var wg sync.WaitGroup

	for i := 0; i < s.numWorkers; i++ {
		wg.Add(1)
		go s.crawlWorker(&wg)
	}
	defer wg.Wait()
}

func (s *Scheduler) crawlWorker(wg *sync.WaitGroup){

	for i := 0; i < s.CQueue.Len(); i++ {

		task := s.CQueue.Fetch()
		ok := s.CStorage.Contains(task)

		if !ok {
			// TODO Logging instead of STDOUT
			fmt.Println("[Info] Crawling : " + task)

			result, err := fetcher.FetchURL(task, false, s.CQueue, s.CStorage)

			if err != nil {
				fmt.Println("[Warning] :", err)
				continue
			}

			siteMetadata := s.fetcherResultToMap(result)
			s.CStorage.Add(task, siteMetadata)
		}
	}
	defer wg.Done()
}

func (s *Scheduler) fetcherResultToMap(fetcherResult fetcher.Result) (map[string][]string){
	siteMetadata := make(map[string][]string)
	siteMetadata["images"] = fetcherResult.Images
	siteMetadata["styles"] = fetcherResult.Styles
	siteMetadata["scripts"] = fetcherResult.Scripts
	return siteMetadata
}
