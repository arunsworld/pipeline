# Data Pipeline APIs functional style with transparent concurrency support

## Concepts

* `Pipeline` - allows for the definition of a data pipeline with a sequence of PreFilters, Transformers and PostFilters
* `Components` - utility structure to define the pipeline components. Including concurrency.
* `FilterOperation` - allows determination of whether an element should stay or be filtered out during pipeline processing
* `TransformOperation` - transforms element. additionally supports filtering. additionally supports returning errors
* `FoldOperation` - folds / reduces elements down

## Pipeline

The Pipeline is the overarching concept that provides the executable behviours of:

* `Apply` - takes a slice and applies the data pipeline to it
* `ApplyAndFold` - applies data pipelien to slice and then folds it down to a single element
* `Stream` - continuously streams data through the pipeline until closure

## Concurrency

* Set the `Concurrency` in `Components` to greater than 1 to get transparent concurrency when processing the data pipeline. 
* Data is the slice/channel is distributed between a number of independent jobs that execute the data pipeline across the filters, transforms and even the folding operation concurrently.
* During streaming this may naturally result in out of sequence data handling, so beware!
* During slice processing (`Apply`) the result is sorted by original sequence allowing sequence to be maintained.
* Goes without saying that when using concurrency all functions passed in must be thread-safe.

### What should Concurrency be set to?

* Concurrency should either be 1 for serial processing.
* Concurrency should generally be set to a reasonably large number (say 100) for parallel processing.
* This parameter does NOT need to correspond to the number of cores because this parameter controls goroutines which can be fairly large and still be effeciently scheduled on the actual hardware available.
* Consideration of the right number might more be linked to memory usage. If each element takes 10MB data to process through the pipeline and concurrency is 100, is that OK? (1GB memory)
* In summary, make the Concurrency fairly high compared to the number of cores but keep it reasonable to avoid memory issues.