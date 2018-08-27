package main

import(
	"fmt"
	"math/rand"
	"time"
	"sync"
	//"runtime"
)

var wg sync.WaitGroup

func UpdateStoreIn2Min(){
	defer wg.Done()
	
	var a []int
	
	for i:=0;i<30;i++{
	 a=append(a,i)
	}
	
	t := time.Now()
	rounded := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute()+1, 0, 0, t.Location())
	
	fmt.Println(rounded.Sub(t))
	 
	time.AfterFunc(rounded.Sub(t),func() {
	 
	fmt.Println(time.Now())
	 
	ticker := time.NewTicker(time.Millisecond * 500)
	 
	defer ticker.Stop()
	done := make(chan bool)
	 
	go func() {
		time.Sleep(10 * time.Second)
		done <- true
	}()
	for {
		select {
		case <-done:
			fmt.Println("Done!")
			return
		case ti := <-ticker.C:
			rand.Shuffle(len(a), func(i, j int) {
			a[i], a[j] = a[j], a[i]
			})
			fmt.Println(a,ti)
		}
	}
})
	
	
}

func main(){
	//runtime.GOMAXPROCS(2)
	rand.Seed(time.Now().UnixNano())
	wg.Add(1)

	go UpdateStoreIn2Min()

	wg.Wait()
    time.Sleep(71 * time.Second)
}