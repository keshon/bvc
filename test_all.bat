@echo off
echo === Running repo tests ===
go test -coverpkg=./internal/repo/... ./internal/tests/repo -v -cover
echo.

echo === Running block tests ===
go test -coverpkg=./internal/storage/block ./internal/tests/storage_block -v -cover
echo.

echo === Running file tests ===
go test -coverpkg=./internal/storage/file ./internal/tests/storage_file -v -cover
echo.

echo === Running snapshot tests ===
go test -coverpkg=./internal/storage/snapshot ./internal/tests/storage_snapshot -v -cover
echo.
