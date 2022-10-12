# Data Pipeline API with transparent concurrency support

## Concepts

* `Pipeline` - define a data pipeline with a sequence of Filters and Transforms
* `Filter` - a function that given an element says whether it should be kept or not
* `Transform` - a function that translates an element from one value to another
* `FoldOperation` - folds / reduces elements down

## Pipeline

The Pipeline is the overarching concept that provides the executable behviours of:

* `Apply` - takes a slice and applies the data pipeline to it
* `ApplyAndFold` - applies data pipeline to a slice and then folds it down to a single element
* `Stream` - continuously streams data through the pipeline until closure

## Concurrency

* Call `Concurrency()` during pipeline creation with a number greater than 1 to get transparent concurrency when processing the data pipeline. 
* Elements will be distributed between a number of independent jobs that execute the data pipeline across the filters, transforms and even the fold operations concurrently.
* During streaming this may naturally result in out of sequence data at the output, so beware!
* During slice processing (`Apply`) the result is sorted by original sequence allowing sequence to be maintained.
* Goes without saying that when using concurrency all functions passed in for Filter and Transform must be thread-safe.

### What should Concurrency be set to?

* Concurrency should be 1 if desiring serial processing.
* Concurrency should generally be set to a reasonably large number (say 100) for parallel processing.
* This parameter does NOT need to correspond to the number of cores because this parameter controls goroutines which can be fairly large and still be effeciently scheduled on the actual hardware cores available.
* Consideration for the right number should likely be more linked to memory or any other resource the processing might be using. If each element takes 10MB data to process through the pipeline and concurrency is 100, 1GB of memory will be used. This is the sort of analysis that should be involved.
* In summary, make the Concurrency fairly high compared to the number of cores but keep it reasonable to avoid memory or resource issues.

## Usage

* For a pipeline that changes 2 filters, 2 transforms and a final filter
* Note: a trasnform may return an error that will stop the execution at the first error encountered

```go
import (
    p "github.com/arunsworld/pipeline"
)

pl := p.New[int]().
        Filter(filterA).
        Filter(filterB).
        Transform(transformA).
        MustTransform(transformB).
        Filter(filterC)

output, err := pl.Apply(data)

```