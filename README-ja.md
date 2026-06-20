<!-- markdownlint-disable MD033 -->
<p align="center">
  <img src="./assets/logo.jpeg" alt="logo" width="400">
</p>

<div align="center">
  <a href="https://github.com/zawa-kyo/pouch/releases/latest"><img src="https://img.shields.io/github/v/release/zawa-kyo/pouch" alt="release"></a>
  <a href="https://github.com/zawa-kyo/pouch?tab=MIT-1-ov-file"><img src="https://img.shields.io/github/license/zawa-kyo/pouch" alt="license"></a>
  <a href="https://github.com/zawa-kyo/pouch/actions/workflows/ci.yml"><img src="https://github.com/zawa-kyo/pouch/actions/workflows/ci.yml/badge.svg?branch=main" alt="ci"></a>
  <a href="https://goreportcard.com/report/github.com/zawa-kyo/pouch"><img src="https://goreportcard.com/badge/github.com/zawa-kyo/pouch" alt="report"></a>
  <a href="https://github.com/zawa-kyo/pouch/tree/main"><img src="https://img.shields.io/github/repo-size/zawa-kyo/pouch" alt="size"></a>
</div>
<!-- markdownlint-enable MD033 -->

# 👜 pouch

`pouch` は、CLI で受け取ったパスからファイルやディレクトリを作成します。
足りないパスは作成し、既存ファイルはそのまま残します。

作りたいパスは決まっているものの、毎回 `mkdir -p` と `touch` のどちらを書くか考えたくない。`pouch` は、そういう場面に絞った小さな CLI です。

## なぜ `pouch` なのか

パスを作るだけでも、通常は複数のコマンドを使い分けることになります。

```sh
mkdir -p notes
mkdir -p src && touch src/main.go
```

`pouch` という名前は、ディレクトリには `mkdir -p`、ファイルには `touch` を使い分ける感覚から付けています。`pouch` は、その 2 つの習慣を 1 つの小さなコマンドにまとめます。パスを渡すと、次のシンプルなルールで自然な結果を選びます。

| パスの形                      | 判定結果               |
| ----------------------------- | ---------------------- |
| 末尾が `/` で終わる           | ディレクトリとして扱う |
| 最後のセグメントに `.` がある | ファイルとして扱う     |
| それ以外                      | ディレクトリとして扱う |

このルールは意図的に小さくしています。`pouch` はスキャフォールディングツールではなく、指定されたパスを作るためのツールです。

## 例

```sh
pouch foo
pouch bar/baz.go
pouch src/main.go test
```

| コマンド                 | 動作                                               |
| ------------------------ | -------------------------------------------------- |
| `pouch foo`              | `foo` ディレクトリを作成する                       |
| `pouch bar/baz.go`       | 親ディレクトリを作成してから `baz.go` を作成する   |
| `pouch src/main.go test` | 各パスを入力順に処理する                           |

実際の動きは次のとおりです。

<!-- markdownlint-disable MD033 -->
<p align="center">
  <img src="./assets/demo.gif" alt="demo" width="640">
</p>
<!-- markdownlint-enable MD033 -->

## インストール

### Go

```sh
go install github.com/zawa-kyo/pouch/cmd/pouch@latest
```

### mise

GitHub backend を直接利用できます。

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

## `mkdir -p` と `touch` との違い

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

このルールで意図どおりに判定できるなら、auto モードのままで十分です。そうでない場合だけ `--file`、`--dir`、または `--mode` で明示します。

## ファイルかディレクトリかを明示する場面

auto ルールでは、判定が分かれやすい名前もあります。

| パス           | auto モードでの判定 | よく使う上書き |
| -------------- | ------------------- | -------------- |
| `Dockerfile`   | ディレクトリ        | `--file`       |
| `Makefile`     | ディレクトリ        | `--file`       |
| `dir.with.dot` | ファイル            | `--dir`        |

> [!IMPORTANT]
> `Dockerfile` と `Makefile` は、auto モードではディレクトリとして扱われます。ファイルとして作成したい場合は `--file` を使ってください。

別の結果にしたい場合は `--file` または `--dir` を使います。

```sh
pouch Dockerfile --file
pouch dir.with.dot --dir
```

モード指定はコマンド全体に効きます。たとえば `pouch Dockerfile test --file` は、`Dockerfile` と `test` の両方をファイルとして扱います。

`--mode file` と `--mode dir` も引き続き使えます。同じコマンド内で `--mode` と `--file` または `--dir` は併用しないでください。

パスが `/` で終わっている場合、file モードはファイルを作らずエラーにします。

## `--strict` を使う場面

デフォルトの `pouch` は、同じコマンドを再実行しやすい挙動です。対象パスがすでに同じ種類で存在していれば、変更せず成功として扱います。

既存パスを失敗扱いにしたい場合は、`--strict` を使います。

```sh
pouch src/main.go test --strict
```

## 振る舞い

| 対象         | 動作                                                   |
| ------------ | ------------------------------------------------------ |
| ファイル     | 足りない親ディレクトリを作成する                       |
| ファイル     | ファイルが存在しなければ作成する                       |
| ファイル     | デフォルトでは、すでに存在するファイルは変更しない     |
| ファイル     | 対象パスがディレクトリとして存在する場合はエラーにする |
| ディレクトリ | `mkdir -p` 相当でディレクトリを作成する                |
| ディレクトリ | すでに存在する場合も成功として扱う                     |
| ディレクトリ | 対象パスがファイルとして存在する場合はエラーにする     |

## CLI

基本形は次のとおりです。

```sh
pouch [flags] PATH...
```

フラグは `PATH...` の前後どちらにも置けます。パス自体が `-` で始まる場合は `--` を使います。

| フラグ                           | 意味                                                       |
| -------------------------------- | ---------------------------------------------------------- |
| `-h`, `--help`                   | ヘルプを表示する                                           |
| `--file`                         | 各パスをファイルとして扱う                                 |
| `--dir`                          | 各パスをディレクトリとして扱う                             |
| `-m`, `--mode <auto\|file\|dir>` | ファイルかディレクトリかを明示する                         |
| `-n`, `--dry-run`                | ファイルシステムを変更せず、予定している操作だけを表示する |
| `-s`, `--strict`                 | 対象パスがすでに存在する場合は失敗させる                   |
| `-v`, `--version`                | バージョンを表示する                                       |
| `-V`, `--verbose`                | 各操作を入力順に表示する                                   |

## 終了挙動

- すべて成功したら終了コード `0` を返す。
- 最初のエラーで非ゼロ終了する。
- エラーは標準エラー出力へ書く。
- 複数パスを受け取った場合も入力順に処理し、最初のエラーで停止する。

## スコープ

`pouch` は意図的に対象を絞っています。

- 対応環境: macOS と Linux に絞る。
- 責務: CLI で受け取ったパスから、ファイルかディレクトリを作成する。プロジェクト構成やファイル内容までは扱わない。
- 判定: 自動判定は小さなルールセットに保ち、明示的に切り替えたい場合だけ `--file`、`--dir`、または `--mode` を使う。
- 使い勝手: 毎回の実行結果が読みやすく予測できる CLI に保つ。対話プロンプトには頼らない。
- 設定: 挙動は各コマンド呼び出しで完結させ、設定ファイルには依存しない。
