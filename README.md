# distributedgrep

An implementation of the [Google MapReduce paper](http://static.googleusercontent.com/media/research.google.com/en//archive/mapreduce-osdi04.pdf), extended to include grep-like functionality. This is implemented using mulitple processes on a single machine but could be extended to use multiple machines. The general idea is that we split the searching of files across multiple processes using the communication protocol that is defined in the paper.


## Usage
```
Usage: ./dgrep pattern [inputfiles]...
Example: ./dgrep "printf" pg-*.txt
```


## Implementation Details
<p align="center">
  <img src="https://user-images.githubusercontent.com/44622454/198502246-6dcbe697-f316-4eda-95ae-265acbf4a49f.png" />
</p>

The grep implementation of MapReduce looks similar to the diagram above. The nice thing about this programming model is the ability to plug in different ways of reading and processing the data as the underlying communication between processes remains the same. This program was originally written as a word count but was easily modified to implement file reading capabilities. 

The general idea is that we have a "master" program that facilitates the distribution and storage of the Map and Reduce tasks. We start up a bunch of Worker processes which communicate with the Master task via RPC. The workers send requests to the master when they need more work as well as when they have finished a task. After giving out a task, the master starts a timer and if the worker has not reported back by the end of that timer the task will be added back to the queue and the next worker to request will receive it. Once a worker completes a Map task it will write a file to disk with the format of `mr-<map task id>-<reduce task id>`, we then report the filename and path to the master program. Once all Map tasks have been completed we use the intermediate files to produce the final output using the Reduce function.
