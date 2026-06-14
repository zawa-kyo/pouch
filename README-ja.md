# pouch

`pouch` は、パスから振る舞いを決めるシンプルな `touch` コマンドです。

1 つのパスを受け取り、ディレクトリかファイルのどちらかを作成します。

- 最後のパスセグメントにドット (`.`) が含まれていれば、ファイルパスとして扱います。
- そうでなければ、ディレクトリパスとして扱います。

ファイルを作るときは、足りない親ディレクトリもあわせて作成します。

## 目的

パスを作るだけでも、通常は複数のコマンドを使い分けることになります。

```sh
mkdir -p sample
mkdir -p "$(dirname sample/temp.ts)" && touch sample/temp.ts
touch sample.ts
```

`pouch` は、それを 1 つのルールにまとめます。

```sh
pouch sample
pouch sample.ts
pouch sample/temp.ts
```

このコマンドは、小さく、役割が明確な道具として設計します。

## 振る舞い

### 自動判定

`pouch` は最後のパスセグメントだけを見ます。

例:

- `sample` -> ディレクトリ
- `sample.ts` -> ファイル
- `sample/temp.ts` -> ファイル
- `.env` -> ファイル

### ファイルとして扱う場合

ファイルと判定した場合の動作:

1. `mkdir -p` 相当で親ディレクトリを作成する。
2. ファイルが存在しなければ作成する。
3. すでに存在する場合は、`touch` のように更新時刻を更新する。

### ディレクトリとして扱う場合

ディレクトリと判定した場合の動作:

1. `mkdir -p` 相当でディレクトリを作成する。
2. すでに存在していれば成功として扱う。

## 対象外

`pouch` は、あらゆるパス作成の問題を解こうとはしません。

自動判定の対象外:

- 拡張子のないファイルを自動で判定すること。
- 内容や MIME からファイル種別を推測すること。
- テンプレートや雛形を生成すること。
- 言語ごとのスキャフォールディング。
- プロジェクト初期化。

## 既知の制約

自動判定ルールは意図的に単純であり、そのぶん割り切りがあります。

例:

- `Dockerfile` は auto モードではディレクトリとして扱われる
- `Makefile` は auto モードではディレクトリとして扱われる
- `dir.with.dot` は auto モードではファイルとして扱われる

こうしたケースは、明示的なモード指定で扱う前提です。

## CLI

### 基本形

```sh
pouch PATH...
```

例:

```sh
pouch sample
pouch sample.ts
pouch sample/temp.ts
pouch src/index.ts test/index.test.ts docs
```

### 想定フラグ

```sh
pouch [flags] PATH...
```

フラグ:

- `-m, --mode <auto|file|dir>`
  自動判定を使わず、ファイルかディレクトリかを明示する。
- `-n, --dry-run`
  ファイルシステムは変更せず、予定している操作だけを表示する。
- `-v, --verbose`
  実行した操作、または実行予定の操作を順番に表示する。
- `--no-touch`
  既存ファイルがある場合に更新時刻を変更しない。
- `-h, --help`
  ヘルプを表示する。
- `--version`
  バージョンを表示する。

### 明示モードの例

```sh
pouch --mode file Dockerfile
pouch --mode dir dir.with.dot
```

## 終了コード

初期リリースでは、次の挙動を想定します。

- すべて成功したら終了コード `0`
- 最初のエラーで非ゼロ終了
- エラーメッセージは標準エラー出力に書く

将来的には、次のような拡張もありえます。

- `--continue-on-error`

ただし、明確な必要が出るまでは初回リリースに含めません。

## 例

### ディレクトリを作る

```sh
pouch sample
```

次のコマンドと同じ意図です。

```sh
mkdir -p sample
```

### カレントディレクトリにファイルを作る

```sh
pouch sample.ts
```

次のコマンドと同じ意図です。

```sh
touch sample.ts
```

### 親ディレクトリごとファイルを作る

```sh
pouch sample/temp.ts
```

次のコマンドと同じ意図です。

```sh
mkdir -p sample
touch sample/temp.ts
```

### 拡張子のないファイルを明示的に作る

```sh
pouch --mode file Dockerfile
```

### ドットを含むディレクトリを明示的に作る

```sh
pouch --mode dir dir.with.dot
```

## 設計方針

`pouch` は小さく保ちます。

中核となる方針:

- 役割は 1 つに絞る: パスを作る。
- 賢い推測より、予測しやすいルールを優先する。
- auto モードは単純なまま保ち、仕様として明示する。
- 曖昧なケースには明示的な上書き手段を用意する。
- 妥当な範囲で標準的なファイルシステムの挙動に合わせる。
- CLI と Go パッケージの両方で使いやすくする。

## 位置づけ

短い説明:

> `pouch` is a path-aware `touch` command.

少し長い説明:

> `pouch` creates directories or files from path-like arguments using a simple, explicit detection rule.
