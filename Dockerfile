FROM gcr.io/distroless/static-debian11:nonroot
ENTRYPOINT ["/baton-splunk"]
COPY baton-splunk /