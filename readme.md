# Great Cabin

A harmonious way to interact with Kubernetes. Several innovative ideas on the road map to make it state-of-the-art, such as:
- [ ] Events timeline - a unique feature for clear root cause analysis.
- [ ] Automatically generated components diagram - a first-of-its-kind feature that lets you know what's in your cluster.
- [ ] Custom home dashboards for different use cases - open GC, see what interests you 
- [ ] Themes - build your own - another unique feature.
- [x] 100% client-side, meaning nothing is installed on the cluster, non-invasive, increasing the adoption rate. You see / have the right to exactly what rights you have in k8s (in this team you will start using the term k8s a lot) 
- [ ] Experimental - Learning elements using ChatGPT and dynamic popups to explain, generate details about alerts, events, objects 
- Technologies: 
    - Backend: Golang / websockets / HTTP 
    - Frontend: plain Html/CSS/JS - nothing fancy. A few libraries used: Datatables.js, d2 charts, bootstrap 5. 


## Pods tab

- are there pods in a state different from  `Running` and `Succeeded` ?
    - Are they Failed - check events
        - Yes. Is the number of replicas the same as desired
        - No
            - Does the service have traffic 
    - Are they pending - check events
        - 

## Services tab
- rank all the services, listing 
