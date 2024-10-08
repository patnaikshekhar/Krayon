curl -X POST \
    -H "content-type:multipart/form-data" \
    -F "name=readme" \
    -F "description=Improves upon your readme" \
    -F "version=v0.0.1" \
    -F binary_mac=@sample_plugins/readme/readme \
    -F binary_linux=@sample_plugins/readme/readme \
    -F binary_windows=@sample_plugins/readme/readme \
    http://localhost:8000/api/v1/plugins


curl -v -X GET \
    http://localhost:8000/api/v1/plugins

curl -v -X GET \
    http://localhost:8000/api/v1/plugins/1/versions

curl -v -X GET \
    http://localhost:8000/api/v1/plugins/test/versions/v0.0.02/platforms/mac -O