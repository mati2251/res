# Remote Script Executor
This is a simple HTTP service for execute scripts in apptainer.
## API
### Table
| URI | GET | POST | PUT | PATCH | DELETE |
| --- | --- | --- | --- | --- | --- |
| `/images` | Retrieve a list of all registered apptainer images (properties only) | Register a new image entry (only name, ID — no binary upload yet) | ❌| ❌| ❌|
| `/images/{name}` | Get properties about the image | ❌ | ❌ | ❌ | Delete the image |
| `/images/{name}/raw` | Get the raw apptainer image file | ❌ | Upload or replace the raw apptainer image file | ❌ | ❌ |
| `/images/{name}/properties` | Get properties about the image | ❌| ❌| ❌| ❌|
| `/jobs` | Retrive a list of jobs | Prepare a new job | ❌| ❌| ❌|
| `/jobs/{id}` | Get job properties | ❌| Put job properties | ❌| ❌|
| `/jobs/{id}/state` | Get job state | ❌| Change job script | ❌| ❌|
| `/jobs/{id}/script` | Get job script | ❌| Put job script | ❌| ❌|
| `/jobs/{id}/log` | Get job logs | ❌| ❌ | ❌| ❌|
| `/jobs/{id}/artifacts` | Get list of job artifacts properties | ❌| ❌ | ❌| Delete job artifacts |
| `/jobs/{id}/artifacts/data` | Get job artifacts data | ❌| ❌ | ❌| ❌|
| `/pipelines` | ❌| Register a new pipeline | ❌| ❌| ❌|
