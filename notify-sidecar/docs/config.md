# Setup

In order to make use of the notify sidecar some preparations need to be taken. Primarily in order to configure the location of the shared volume for a java application to write information to and the location of the central heap dump service.

## Kubernetes

As the name suggest the notify sidecar should be ran as a sidecar next to the actual application. This deployment will give a minimal example how it should look like. Please take note on the shared volume and the java opts to ensure the jvm will write to the correct location in case a OOMException is thrown.

```yaml
apiVersion: apps/v1
kind: Deployment
spec:
  selector:
    matchLabels:
      app.kubernetes.io/instance: java-app
      app.kubernetes.io/name: java-app
  template:
    spec:
      containers:
        - name: java-app
          image: example-java-application
          env:
            - name: JAVA_TOOL_OPTIONS
              value: -Xms512m -Xmx512m -XX:+HeapDumpOnOutOfMemoryError -XX:HeapDumpPath=/heap-dumps/
          resources:
            limits:
              memory: 1Gi # needs to be higher than maximum heap 
            requests:
              memory: 1Gi
          volumeMounts:
          - mountPath: /heap-dumps
            name: heap-dumps
        - name: notify-sidecar
          image: ghcr.io/dbschenker/heap-dump-management/notify-sidecar:release-1.1.3-28cc0112
          command:
            - /app
          env:
            - name: APP_CONFIG_FILE
              value: /opt/config.json
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  apiVersion: v1
                  fieldPath: metadata.name
            - name: NOTIFY_SIDECAR_LOG_LEVEL
              value: WARNING
          volumeMounts:
            - mountPath: /heap-dumps
              name: heap-dumps
            - mountPath: /opt
              name: heap-dump-config
      volumes:
        - emptyDir: {}
          name: heap-dumps
        - configMap:
            defaultMode: 420
            name: java-app-heap-dump-cm
          name: heap-dump-config
```

# Configuration

The heap dump service can be configured with a json config and environment variables.  
Here is an example json configuration: 

```json
{
    "metrics": {
        "port": 8081,
        "path": "/metrics"
    },
    "WatchPath": {
        "path": "/heap-dumps"
    },
    "Middleware": {
        "endpoint": "https://test.svc.cluster.local" 
    },
    "ServiceOwner": {
        "tenant": "testTenant"
    }
}
```

this `config.json` file can be referenced by the environment variable `APP_CONFIG_JSON`.  
Other environment variables include: 

```yaml
- name: APP_CONFIG_FILE
  value: /opt/config.json
- name: POD_NAME
  valueFrom:
    fieldRef:
      apiVersion: v1
      fieldPath: metadata.name
- name: NOTIFY_SIDECAR_LOG_LEVEL
  value: WARNING
```
