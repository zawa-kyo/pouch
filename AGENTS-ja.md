# pouch エージェントガイド

## 文書同期

- `README.md` と `README-ja.md` は意味を揃えて保つ。
- `AGENTS.md` と `AGENTS-ja.md` は意味を揃えて保つ。
- 対になる文書の片方を更新する場合、ユーザーが一時的な不一致を明示的に求めていない限り、同じ変更の中でもう片方も更新する。
- `*-ja.md` は逐語訳ではなく、自然な日本語で書く。

## 目的

この文書は、`pouch` の内部設計と実装方針をまとめたものです。
`README.md` にある外部仕様をもとに、最初の Go 実装を進めるときの基準として使います。

## プロダクトの境界

- `pouch` の中心は CLI。
- 再利用可能な Go パッケージは、CLI とテストを支えるために存在する。
- このプロジェクトを汎用スキャフォールディングツールへ広げない。
- 実装の焦点は、あくまでパス作成に置く。

## 中核の挙動

- auto モードでは、最後のパスセグメントだけを見る。
- 最後のセグメントにドットが含まれていれば、ファイルとして扱う。
- そうでなければ、ディレクトリとして扱う。
- ファイルとして扱う場合は、先に足りない親ディレクトリを作る。
- ディレクトリとして扱う場合は、`mkdir -p` 相当で作成する。

## 対象外

- テンプレート生成
- ファイル内容の生成
- MIME や言語の判定
- プロジェクト初期化ワークフロー
- 初回リリースでの設定ファイル対応
- 初回リリースでの対話モード

## API 方針

公開 API は小さく保つ。

想定パッケージ:

```go
package pouch
```

想定 API:

```go
package pouch

import "os"

type Mode int

const (
    ModeAuto Mode = iota
    ModeFile
    ModeDir
)

type Kind int

const (
    KindFile Kind = iota
    KindDir
)

type Options struct {
    Mode     Mode
    DirPerm  os.FileMode
    FilePerm os.FileMode
    Touch    bool
    DryRun   bool
}

type Result struct {
    Path    string
    Kind    Kind
    Created bool
    Touched bool
}

func Detect(path string) Kind
func Create(path string, opts Options) (Result, error)
func CreateMany(paths []string, opts Options) ([]Result, error)
```

補足:

- `Detect` は決定的で、副作用を持たないようにする。
- 振る舞いの中心は `Create` に置く。
- `CreateMany` は入力順を保つ。

## 判定ルール

基準ロジック:

1. `filepath.Base(path)` を計算する。
2. ベース名にドットが含まれていれば、ファイルを返す。
3. それ以外は、ディレクトリを返す。

重要な例:

- `sample` -> ディレクトリ
- `sample.ts` -> ファイル
- `sample/temp.ts` -> ファイル
- `.env` -> ファイル
- `Dockerfile` -> auto モードではディレクトリ
- `dir.with.dot` -> auto モードではファイル

このロジックは単純なまま保つ。外部仕様が変わらない限り、隠しファイル、既知のファイル名、末尾スラッシュ向けの追加ヒューリスティックは入れない。

## CLI 要件

想定フラグ:

- `--mode auto|file|dir`
- `--dry-run`
- `--verbose`
- `--no-touch`
- `--help`
- `--version`

想定挙動:

- 完全成功なら終了コード `0`
- 最初のエラーで非ゼロ終了
- エラーは標準エラー出力へ書く
- dry-run ではファイルシステムを変更しない
- verbose では入力順に各操作の予定または実行内容を表示する

## ファイルシステム上の挙動

### ファイルモード

- 親ディレクトリは `os.MkdirAll` で作成する。
- ファイルがなければ作成する。
- ファイルが存在し、touch が有効なら時刻を更新する。
- touch が無効なら既存ファイルの時刻は変更しない。

### ディレクトリモード

- ディレクトリは `os.MkdirAll` で作成する。
- すでに存在していても成功として扱う。

## 権限

コード上では明示的な既定値を使う。

- ディレクトリ: `0o755`
- ファイル: `0o644`

実際の有効権限は `umask` の影響を受ける。

## エラーハンドリング

エラーにはパス文脈を付けて包む。

例:

- `detect "sample.ts": ...`
- `create parent directory for "sample/temp.ts": ...`
- `create file "sample.ts": ...`
- `touch file "sample.ts": ...`

初回リリースでは、独自の複雑なエラー階層よりも、平明で直接的なメッセージを優先する。

## パッケージ構成

最初の構成は浅く保つ。

```text
.
├── cmd/
│   └── pouch/
│       └── main.go
├── internal/
│   └── cli/
│       └── flags.go
├── pouch.go
├── create.go
├── detect.go
├── types.go
└── README.md
```

想定役割:

- `types.go`
  公開 enum と option/result 型
- `detect.go`
  判定ロジックだけを置く
- `create.go`
  ファイルシステム操作の本体を置く
- `pouch.go`
  公開エントリポイントと小さな調停処理を置く
- `cmd/pouch/main.go`
  CLI のエントリポイント
- `internal/cli/flags.go`
  CLI フラグの解析と検証

必要性が明確になるまでは、これ以上パッケージを分割しない。

## 実装メモ

### パス処理

- 標準ライブラリの `path/filepath` を使う。
- `mkdir` や `touch` などの OS コマンドは呼ばない。
- 独自の正規化より、標準ライブラリの挙動を優先する。

### ファイル作成

推奨手順:

- まず親ディレクトリを確実に作る。
- `os.OpenFile` を create 可能なフラグ付きで使う。
- ファイルはすぐ閉じる。
- touch が有効な場合は、時刻更新を明示的に行う。

### タイムスタンプ

touch の挙動は意図的に扱う。

- 新規作成されたファイルは `Created` とみなす。
- 既存ファイルの時刻更新は `Touched` とみなす。
- dry-run の結果は、変更せずに意図した操作を反映する。

### 複数パス処理

- `CreateMany` は入力順に処理する。
- v0.1.0 では最初のエラーで停止する。
- 失敗前の成功結果を返す形が CLI にとって有用なら、その形を採用してよい。
- API が不自然になるなら、API 自体は単純に保ち、部分結果の扱いは CLI 層に寄せる。

## テスト方針

見通しがよくなる箇所では、テーブル駆動テストを使う。
ファイルシステムの隔離には `t.TempDir()` を使う。

必要なテスト範囲:

- 判定ロジック
- ディレクトリ作成
- ファイル作成
- 親ディレクトリ作成
- 既存ファイルの touch 挙動
- `--mode file` 上書き
- `--mode dir` 上書き
- ドットファイルの扱い
- `Dockerfile` のような曖昧な名前
- `dir.with.dot` のようなドット入りディレクトリ名
- dry-run 挙動

代表ケース:

- `sample` はディレクトリを作る
- `sample.ts` はファイルを作る
- `sample/temp.ts` は親ディレクトリとファイルを作る
- `.env` はファイルを作る
- auto モードの `Dockerfile` はディレクトリを作る
- file モードの `Dockerfile` はファイルを作る
- auto モードの `dir.with.dot` はファイルを作る
- dir モードの `dir.with.dot` はディレクトリを作る

## リリース計画

### v0.1.0

含めるもの:

- `PATH...` を受け取る CLI
- 自動判定
- `--mode auto|file|dir`
- `--dry-run`
- `--verbose`
- `--no-touch`
- 公開 Go パッケージ
- ユニットテスト
- README

含めないもの:

- シェル補完
- 設定ファイル
- テンプレート生成
- 対話プロンプト
- プラグイン機構

## レビュー観点

マージ前に次を確認する。

- コードは依然として単純な外部ルールを反映しているか
- 公開 API は内部構造より小さいままか
- 曖昧なケースは隠されず文書化されているか
- CLI の出力は予測しやすく、過剰に騒がしくないか
- テストは実装の細部ではなく、観測可能な挙動を見ているか

## 変更の規律

- ヒューリスティックを黙って足さない。
- README を更新せずにスコープを広げない。
- 新しいエッジケースに特別扱いが必要なら、賢い自動判定より明示モードを優先する。
- 中核の性格は保つ。小さく、予測しやすく、パスを理解する `touch` であり続ける。
