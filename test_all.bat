@echo off
echo ### Running repo tests ###
go test -coverpkg=./internal/repo/... ./internal/tests/repo -v -cover
echo.

echo ### Running block tests ###
go test -coverpkg=./internal/store/block ./internal/tests/store_block -v -cover
echo.

echo ### Running file tests ###
go test -coverpkg=./internal/store/file ./internal/tests/store_file -v -cover
echo.

echo ### Running snapshot tests ###
go test -coverpkg=./internal/store/snapshot ./internal/tests/store_snapshot -v -cover
echo.

echo ### Running repotools tests ###
go test -coverpkg=./internal/repotools ./internal/tests/repotools -v -cover
echo.

echo ### Running command tests ###
go test -coverpkg=./internal/command ./internal/tests/command -v -cover
echo.