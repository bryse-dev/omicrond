# omicrond
A  job scheduler daemon to replace cron on Unix systems.  This is achieved by focusing on minimal dependencies, stability, and robust failure handling... the main arguments for sticking with the basic cron utility.

Planned features:
*  Parse crontab for easy migration
*  Jobs can be dependant on completion of other jobs and their return code
*  API accessible for online updates and expansionary tools:
 *  Reachable by unix and/or tcp socket
 *  Output from jobs can be retrieved 
 *  Configuration can be managed:  new jobs added/ existing modified/ daemon settings changed
 *  Job management: view status/ kill running/ start job /test job
*  Jobs can be grouped allowing:
 * staggered start times (group of 5 jobs spaced between start_window and stop_window)
*  NTP integration
