# pouch エージェントガイド

## 文書同期

- `README.md` と `README-ja.md` は意味を揃えて保つ。
- `AGENTS.md` と `AGENTS-ja.md` は意味を揃えて保つ。
- 対になる文書の片方を更新する場合、ユーザーが一時的な不一致を明示的に求めていない限り、同じ変更の中でもう片方も更新する。
- `*-ja.md` は逐語訳ではなく、自然な日本語で書く。

## 目的

この文書は、`README.md` にある仕様をもとに `pouch` の最初の Go 実装を進めるときの基準として使う。

## スコープ

- `pouch` の中心は CLI。
- Go パッケージは CLI とテストを支えるために置く。
- プロジェクトの焦点はパス作成に絞る。
- 汎用スキャフォールディングツールには広げない。

## プロダクトルール

### 判定

- auto モードでは最後のパスセグメントだけを見る。
- 最後のセグメントにドットが含まれていれば、ファイルとして扱う。
- そうでなければ、ディレクトリとして扱う。
- 外部仕様が変わらない限り、隠しファイル、既知のファイル名、末尾スラッシュ向けのヒューリスティックは足さない。

例:

- `sample` -> ディレクトリ。
- `sample.ts` -> ファイル。
- `sample/temp.ts` -> ファイル。
- `.env` -> ファイル。
- `Dockerfile` -> auto モードではディレクトリ。
- `dir.with.dot` -> auto モードではファイル。

### ファイルシステム上の挙動

- ファイルとして扱う場合は、先に足りない親ディレクトリを作る。
- ファイルが存在しなければ作成する。
- 既存ファイルがある場合は、何も変更しない。
- ディレクトリとして扱う場合は、`mkdir -p` 相当で作成する。
- ディレクトリがすでに存在する場合は、成功として扱う。

### 対象外

- テンプレート生成。
- ファイル内容の生成。
- MIME や言語の判定。
- プロジェクト初期化ワークフロー。
- 初回リリースでの設定ファイル。
- 初回リリースでの対話モード。

## CLI 契約

想定フラグ:

- `--mode auto|file|dir`
- `--dry-run`
- `--verbose`
- `--help`
- `--version`

想定挙動:

- 完全成功なら終了コード `0`。
- 最初のエラーで非ゼロ終了。
- エラーは標準エラー出力へ書く。
- dry-run ではファイルシステムを変更しない。
- verbose では入力順に各操作の予定または実行内容を表示する。

## パッケージ API

公開 API は小さく保つ。

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
    DryRun   bool
}

type Result struct {
    Path    string
    Kind    Kind
    Created bool
}

func Detect(path string) Kind
func Create(path string, opts Options) (Result, error)
func CreateMany(paths []string, opts Options) ([]Result, error)
```

補足:

- `Detect` は決定的で、副作用を持たないようにする。
- 振る舞いの中心は `Create` に置く。
- `CreateMany` は入力順を保つ。

## 実装メモ

### パス処理

- 標準ライブラリの `path/filepath` を使う。
- `mkdir` や `touch` などの OS コマンドは呼ばない。
- 独自の正規化より、標準ライブラリの挙動を優先する。

### ファイル作成

- まず親ディレクトリを確実に作る。
- `os.OpenFile` を create 可能なフラグ付きで使う。
- ファイルはすぐ閉じる。
- 既存ファイルがある場合は、タイムスタンプを変えずに成功として返す。

### 権限

コード上では明示的な既定値を使う。

- ディレクトリ: `0o755`。
- ファイル: `0o644`。

実際の有効権限は `umask` の影響を受ける。

### エラー

エラーにはパス文脈を付けて包む。

例:

- `detect "sample.ts": ...`
- `create parent directory for "sample/temp.ts": ...`
- `create file "sample.ts": ...`
- `create directory "sample": ...`

初回リリースでは、独自の複雑なエラー階層より、平明で直接的なメッセージを優先する。

### パッケージ構成

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

役割:

- `types.go`: 公開 enum と option/result 型。
- `detect.go`: 判定ロジックだけを置く。
- `create.go`: ファイルシステム操作の本体を置く。
- `pouch.go`: 公開エントリポイントと小さな調停処理を置く。
- `cmd/pouch/main.go`: CLI のエントリポイント。
- `internal/cli/flags.go`: CLI フラグの解析と検証。

必要性が明確になるまでは、これ以上パッケージを分割しない。

## テスト

- 見通しがよくなる箇所ではテーブル駆動テストを使う。
- ファイルシステムの隔離には `t.TempDir()` を使う。
- 判定ロジック、ファイル作成、ディレクトリ作成、親ディレクトリ作成、明示モード上書き、曖昧な名前、dry-run 挙動をカバーする。

代表ケース:

- `sample` はディレクトリを作る。
- `sample.ts` はファイルを作る。
- `sample/temp.ts` は親ディレクトリとファイルを作る。
- `.env` はファイルを作る。
- auto モードの `Dockerfile` はディレクトリを作る。
- file モードの `Dockerfile` はファイルを作る。
- auto モードの `dir.with.dot` はファイルを作る。
- dir モードの `dir.with.dot` はディレクトリを作る。

## リリース範囲

### v0.1.0

含めるもの:

- `PATH...` を受け取る CLI。
- 自動判定。
- `--mode auto|file|dir`。
- `--dry-run`。
- `--verbose`。
- 公開 Go パッケージ。
- ユニットテスト。
- README。

含めないもの:

- シェル補完。
- 設定ファイル。
- テンプレート生成。
- 対話プロンプト。
- プラグイン機構。

## レビュー観点

- コードは文書化された判定ルールを反映しているか。
- 公開 API は内部構造より小さいままか。
- 曖昧なケースは隠されず文書化されているか。
- CLI の出力は予測しやすく、過剰に騒がしくないか。
- テストは実装の細部ではなく、観測可能な挙動を見ているか。

## 変更の規律

- ヒューリスティックを勝手に追加しない。
- スコープを広げるときは、README もあわせて更新する。
- 自動判定を複雑にするより、明示モードで上書きできる設計を優先する。
- 小さくて予測しやすいパス作成ツールという中核の性格を保つ。
