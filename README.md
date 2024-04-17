# goworker
go worker pool

# cases

`./test/main.go`

a normal case is like:

```
pool, err := goworker.NewPool(20)
if err != nil {
	log.Fatal(err)
}

for i := 1; i <= 500; i++ {
	n := i
	pool.AddTask(func() {
		log.Print(n, n*n)
	})
}

pool.Done()
```

# usage

## calls

```
pool, err := NewPool(20)

err := pool.AddTask(func() {...})

pool.Done()
```

## create worker pool

```
pool, err := goworker.NewPool(20)
```

create a worker pool with 20 workers.

the number of worker should be in area `(0, 10000]`.

## add task

```
err := pool.AddTask(func() {...})
```

a task should be `func() {}` type.

`pool.AddTask` should be called before `pool.Done`, otherwise an error returned.

```
pool, _ := goworker.NewPool(5)
pool.Done()

f := func() {
	log.Print("hi")
}

if err := pool.AddTask(f); err != nil {
	log.Fatal(err) // task queue done, no more task
}
```

## done

```
pool.Done()
```

when you add all your tasks to the queue, you can call `pool.Done` immediately.

the call of `pool.Done` will:

* first accept no more tasks
* then close the task queue channel
* finnally wait for all workers finish their tasks

# others

## task queue

the task queue has a 10 times capacity of worker number.

when the queue is full, you will be waiting at calling `pool.AddTask`.

dont call `pool.Done` when you are waiting to add task to the queue.

# maybe new features

* task timeout
* info from tasks
* more logs
* ...
