//go:build tools

package main

import _ "github.com/matryer/moq"

// 当該ファイルにはgo:generateで実行するGo製のツールを定義する
// これらのツールは、ビルドタグを指定しない本番環境用アプリケーションのビルド時には無視される
