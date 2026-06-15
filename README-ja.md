<!-- markdownlint-disable MD033 -->
<p align="center">
  <img src="./assets/logo.jpeg" alt="logo">
</p>

<div align="center">
  <a href="https://github.com/koki-develop/gat/releases/latest"><img src="https://img.shields.io/github/v/release/zawa-kyo/pouch" alt="release"></a>
  <a href="https://github.com/zawa-kyo/pouch?tab=MIT-1-ov-file"><img src="https://img.shields.io/github/license/zawa-kyo/pouch" alt="license"></a>
  <a href="https://github.com/zawa-kyo/pouch/actions/workflows/ci.yml"><img src="https://github.com/zawa-kyo/pouch/actions/workflows/ci.yml/badge.svg?branch=main" alt="ci"></a>
  <a href="https://goreportcard.com/report/github.com/zawa-kyo/pouch"><img src="https://goreportcard.com/badge/github.com/zawa-kyo/pouch" alt="report"></a>
  <a href="https://github.com/zawa-kyo/pouch/tree/main"><img src="https://img.shields.io/github/repo-size/zawa-kyo/pouch" alt="size"></a>
</div>
<!-- markdownlint-enable MD033 -->

# 👜 pouch

`pouch` は、パスからファイルまたはディレクトリを作成する小さな CLI です。
足りないパスは作成しますが、既存ファイルは変更しません。

auto モードでは、次の小さなルールセットで判定します。

| パスの形                      | 判定結果               |
| ----------------------------- | ---------------------- |
| 末尾が `/` で終わる           | ディレクトリとして扱う |
| 最後のセグメントに `.` がある | ファイルとして扱う     |
| それ以外                      | ディレクトリとして扱う |

このルールにより、`mkdir -p` と `touch` を都度使い分けなくても、よくあるパスを作成できます。

## 例

```sh
pouch notes
pouch notes/today.md
pouch src/main.go test
```

| コマンド                 | 動作                                               |
| ------------------------ | -------------------------------------------------- |
| `pouch notes`            | `notes` ディレクトリを作成する                     |
| `pouch notes/today.md`   | 親ディレクトリを作成してから `today.md` を作成する |
| `pouch src/main.go test` | 各パスを入力順に処理する                           |

## インストール

### Go

```sh
go install github.com/zawa-kyo/pouch/cmd/pouch@latest
```

### mise

まずは GitHub backend を直接使えます。

```sh
mise use -g github:zawa-kyo/pouch@latest
```

短いツール名で使いたい場合は、mise の設定に alias を追加します。

```toml
[tool_alias]
pouch = "github:zawa-kyo/pouch"
```

その後、次のようにインストールして有効化できます。

```sh
mise use -g pouch@latest
```

## 目的

パスを作るだけでも、通常は複数のコマンドを使い分けることになります。

```sh
mkdir -p notes
mkdir -p src && touch src/main.go
```

`pouch` という名前は、ディレクトリには `mkdir -p`、ファイルには `touch` を使い分ける感覚から付けています。`pouch` は、その 2 つの習慣を、パスを受け取る 1 つの小さなコマンドにまとめたものです。

作りたいパスは決まっているものの、毎回 `mkdir -p` と `touch` のどちらを書くか考えたくない。そういう場面に絞って使えるようにしています。

## 関連ツール

`pouch` は、既存の標準ツールを置き換えるというより、その組み合わせを扱いやすくするツールです。

- `mkdir -p` は、必要な親ディレクトリをまとめて作成する。
- `touch` は、対象ファイルが存在しない場合に空ファイルを作成する。

Unix の基本コマンドを明示的に使いたい場面では、そのまま `mkdir -p` や `touch` を使えば十分です。`pouch` は、パスそのものから操作を決めたい場面に絞った CLI です。

## auto モードの判定

auto モードでは、まず末尾が `/` かどうかを見ます。そうでなければ最後のパスセグメントを見ます。

| パス             | auto モードでの判定 |
| ---------------- | ------------------- |
| `sample`         | ディレクトリ        |
| `sample.ts`      | ファイル            |
| `sample/temp.ts` | ファイル            |
| `.env`           | ファイル            |
| `dir.with.dot/`  | ディレクトリ        |

> [!NOTE]
> `pouch` は、このルールを意図的に小さく保っています。既知のファイル名や MIME type から意図を推測しません。auto モードでディレクトリを明示できる手掛かりは、末尾のスラッシュだけです。

## `--mode` を使う場面

auto ルールでは、判定が分かれやすい名前もあります。

| パス           | auto モードでの判定 | よく使う上書き |
| -------------- | ------------------- | -------------- |
| `Dockerfile`   | ディレクトリ        | `--mode file`  |
| `Makefile`     | ディレクトリ        | `--mode file`  |
| `dir.with.dot` | ファイル            | `--mode dir`   |

> [!IMPORTANT]
> `Dockerfile` と `Makefile` は、auto モードではディレクトリとして扱われます。ファイルとして作成したい場合は `--mode file` を使ってください。

別の結果にしたい場合は `--mode` を使います。

```sh
pouch --mode file Dockerfile
pouch --mode dir dir.with.dot
```

パスが `/` で終わっている場合、`--mode file` はファイルを作らずエラーにします。

## 振る舞い

| 対象         | 動作                                                   |
| ------------ | ------------------------------------------------------ |
| ファイル     | 足りない親ディレクトリを作成する                       |
| ファイル     | ファイルが存在しなければ作成する                       |
| ファイル     | すでに存在するファイルは変更しない                     |
| ファイル     | 対象パスがディレクトリとして存在する場合はエラーにする |
| ディレクトリ | `mkdir -p` 相当でディレクトリを作成する                |
| ディレクトリ | すでに存在する場合も成功として扱う                     |
| ディレクトリ | 対象パスがファイルとして存在する場合はエラーにする     |

## CLI

基本形は次のとおりです。

```sh
pouch [flags] PATH...
```

| フラグ                           | 意味                                                       |
| -------------------------------- | ---------------------------------------------------------- |
| `-h`, `--help`                   | ヘルプを表示する                                           |
| `-m`, `--mode <auto\|file\|dir>` | ファイルかディレクトリかを明示する                         |
| `-n`, `--dry-run`                | ファイルシステムを変更せず、予定している操作だけを表示する |
| `-v`, `--version`                | バージョンを表示する                                       |
| `-V`, `--verbose`                | 各操作を入力順に表示する                                   |

## 終了挙動

- すべて成功したら終了コード `0` を返す。
- 最初のエラーで非ゼロ終了する。
- エラーは標準エラー出力へ書く。
- 複数パスを受け取った場合も入力順に処理し、最初のエラーで停止する。

## スコープ

`pouch` は意図的に対象を絞っています。

| v0.1 に含めるもの      | v0.1 に含めないもの                |
| ---------------------- | ---------------------------------- |
| macOS 対応             | Windows 対応                       |
| Linux 対応             | テンプレート生成                   |
| CLI 引数からのパス作成 | ファイル内容の生成                 |
| 単純な自動判定         | プロジェクトスキャフォールディング |
| 明示モードによる上書き | 設定ファイル                       |
| 予測しやすい CLI 挙動  | 対話プロンプト                     |
