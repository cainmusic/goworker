# 如何实现一个worker pool

## worker pool的核心

worker pool的核心就是，worker是什么。

当我们有一系列任务，比如有10000张图片需要下载。

我们先实现一个`download(url)`的函数下载一张图片，那么10000张图片可以写作如下：

```
func main() {
	for _, url := range urls {
		download(url)
	}
}
```

相当于我们实现了一个依次处理任务的实体，依次处理了10000个任务。

由于main也是一个goroutine，所以我们实际上实现了一个goroutine来处理这10000个任务。

main这个goroutine本身就是一个worker。

我们可以简单理解为：一个依次处理一系列任务的goroutine就是一个worker。

## 多一些worker就可以解决问题吗

看完前面的例子，我很担心main的健康状况。

main这个worker太累了，要做10000个任务，为了worker的健康，我们多引入一些worker。

```
func main() {
	for _, url := range urls {
		go download(url)
	}
}
```

因为有10000个url，即10000个下载任务，我们循环启动了10000个goroutine，即10000个worker。

但这些worker每一个仅处理一个下载便会退出。

如果接下来我们有1000000（一百万）个任务，或者100000000（一亿）个任务，或者更多，我们也要启动这么多worker吗？

如果你的计算机可以撑得住，你可以这样做，但现实并不如此。

即使你的机器很厉害，可以同时启动一亿个任务，但如果你同时进行这些下载请求，对方可能会认为你在攻击他。

我们最终的目的是在可接受的时间范围内，用可接受的计算机资源，完成这些任务。

## 任务拆解

一个worker下载一张图片可能要一秒钟，但你并不需要在一秒内完成所有任务，也许你期待在五分钟内完成这些任务就可以了。

于是你可以制造100个worker，每个worker下载100张图片，这样考虑到下载时间和调度时间，五分钟应该绰绰有余了。

这样每个worker也不会很疲惫，显得你是个不错的领导。

你实现了一个函数`download100pics(urls)`的函数，并且将10000个任务拆解成100个任务组，每个任务组有100个任务。

```
func main() {
	// 每个urls有100个任务
	for _, urls := range urlsGroup {
		go download100pics(urls)
	}
}
```

任务拆解的方法有很多，上面是一个常规可以理解的机械化拆解方式，非常适合小学生理解。

就好像：

```
老师想要给班上住院的小朋友送1000个纸鹤
班里有50个小朋友可以叠纸鹤
于是老师让每个小学生晚上回家叠20个，第二天带到学校来
```

这个分配虽然很机械，看起来还挺公平，但对于计算机来说，我们一般不这么做。

计算机又不会喊累，而且我们通常的目的是尽快完成任务。

如果某个小学生做的很慢，甚至家里没纸了，叠不出20个纸鹤，老师想要的结果就无法达成了。

对于计算机来说，还有更符合我们需求的任务拆解方式。

## 任务队列

由于我们的worker是不知疲倦的，而且每个worker工作的时间不等。

还是以前面10000张图片为例，这里有些图片大，有些图片小，有些请求网络好，有些请求网络差。

worker完成每个任务的时间都不完全相等，如果按照机械的拆解方式，可能有些worker早早就完成了，最后大家都在等待慢的worker。

所以我们可以不提前拆解，而是把所有任务放进一个大池子里，只要worker空闲了，就去池子里取任务，取到啥做啥。

这个任务池的实现方式很多，我们这里实现一个用带缓冲的channel实现的任务队列。

其核心代码如下：

```
var taskQueue = make(chan task, 200)

func worker() {
	for {
		task := <- taskQueue
		do(task)
	}
}

func main() {
	for i := 0; i < 100; i++ {
		go worker
	}

	for _, task := taskList {
		taskQueue <- task
	}

	for {}
}
```

如此，我们启动了100个worker，这些worker会去从taskQueue里取任务来做，直到完成所有任务。

在这些过程中，每个worker完成的任务数不尽相同，完成的顺序也不固定，不过目前我们还不在意这些。

## worker的启动和退出

上面的代码实际上已经可以完成任务了，但程序并不完善，不会自动退出，你得手动控制其结束。

而且是不论main goroutine还是其他worker goroutine，都无法自动退出。

所以，我们需要对各个goroutine的退出进行控制。

控制goroutine的方法很多，一般常用channel+select，这里的channel通常是worker用来接收结束信号的。

比如：

```
func worker() {
	for {
		select {
		case task := <- taskQueue:
			do(task)
		case <- doneChan:
			return
		}
	}
}
```

这个worker有一个隐含的约定条件，控制方向doneChan发送信号的时候，需要确保taskQueue为空或者说任务都做完了。

否则会出现任务并未完成就退出worker的情况。

我们的实现并不如此，我们在实现中用到了一个go语言的range channel的特性。

即：

* 多个goroutine可以用range操作同一个channel
* range会一直从channel中获取信息，直到channel被关闭并排空（注意理解：被关闭并排空）

代码如下：

```
var taskQueue = make(chan task, 200)

func worker() {
	for task := range taskQueue{
		do(task)
	}
}

func main() {
	for i := 0; i < 100; i++ {
		go worker
	}

	for _, task := taskList {
		taskQueue <- task
	}

	close(taskQueue)

	for {}
}
```

如代码所示，在main goroutine执行了`close(taskQueue)`后，worker会继续把taskQueue中的任务做完，之后退出。

至此，worker的启动和退出都完成了。

## main goroutine的退出

最后就剩下main goroutine的退出了。

我们需要在worker退出的时候告诉main goroutine这个信息。

一个简单的实现就是sync包里的sync.WaitGroup了。

最终，代码如下：

```
var wg sync.WaitGroup
var taskQueue = make(chan task, 200)

func worker() {
	defer wg.Done()
	for task := range taskQueue{
		do(task)
	}
}

func main() {
	wg.Add(100)

	for i := 0; i < 100; i++ {
		go worker
	}

	for _, task := taskList {
		taskQueue <- task
	}

	close(taskQueue)

	wg.Wait()
}
```

上述代码的大体控制流程如下：

* `wg.Add(100)`设置需要等待完成的worker数
* 发送全部任务后`close(taskQueue)`通知worker全部任务发送完成
* worker完成全部任务后退出并`defer wg.Done()`告知main goroutine该worker完成
* `wg.Wait()`等待全部worker完成后退出程序

至此，全部任务完成，并且程序正常退出。

# 其他实现

## 对worker进行统一调度

前面的实现用的是worker去taskQueue里抢任务。

还有一种实现是调度者在拿到任务的时候，去worker pool里查看哪个worker是空闲的，并直接将任务分配给这个worker。

这里面也会有不同的调度策略，比如遇到谁空闲就给谁，或者谁空闲时间长就给谁，等等。

甚至，如果调度者掌握了各个worker的强度，还可以对worker进行更细致的分配。

## 任务超时

有时候某些任务会遇到各种问题无法完成，我们可以接受部分任务没有正常完成，但不应该使这些任务卡住我们的主程序。

因而可以在任务中增加控制代码，在任务超时的时候主动取消任务的继续执行，使任务进程继续执行其他待处理的任务。

## 任务报告

由于任务完成时可能会有各种问题，包括超时等，会产生报错，或者任务正常完成也可能会产生结果。

如何统一的处理这些结果，也是可以实现的功能。

## 复杂的任务池

有些任务会有前后依赖关系，需要前置任务完成了才能执行后面的任务。

我们可以在业务逻辑中实现这样的功能，也可以改造任务池使worker pool本身支持更复杂的任务池的实现。

## 等等

等等。。。
