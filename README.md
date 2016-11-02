# omicrond
A  job scheduler daemon to surpass cron on Unix systems.  This is achieved by focusing on minimal dependencies, stability, and robust failure handling (the main arguments for sticking with the basic cron utility).... Then taking it a step further and adding features based on the first-hand failings of cron in an enterprise environment.  

Planned features:
- [x] Parse crontab for easy migration
- [x] Jobs can be self-locking for safe iterative runs
- [ ] Jobs can be dependant on completion of other jobs and their return code
- [x] API accessible for online updates and expansionary tools:
- [x] API Reachable by tcp socket
- [ ] API Reachable by unix socket
- [ ] API output from jobs can be retrieved 
- [x] API configuration can be managed:  new jobs added/ existing modified/ daemon settings changed
- [ ] Job management: view status/ kill running/ start job /test job
- [ ] Jobs can be grouped allowing staggered start times (group of 5 jobs spaced between start_window and stop_window)
- [ ] NTP integration

