# Claude Code Session Viewer

Claude Codeのセッション履歴を3カラムレイアウトで見やすく表示するWebアプリケーションです。

## 機能

- **3カラムレイアウト**: プロジェクト→セッション→プロンプト→回答の流れで直感的に閲覧
- **左サイドバー**: プロジェクトとセッションを階層表示
- **中央サイドバー**: 選択したセッションのプロンプト一覧
- **右メインエリア**: 選択したプロンプトに対するAIの回答を表示
- **検索機能**: プロジェクト名での絞り込み検索

## 必要要件

- Go 1.21以上
- Claude Code CLIがインストール済み（`~/.claude/`ディレクトリが存在すること）

## インストール

```bash
# リポジトリをクローン
git clone https://github.com/yugo-ibuki/claude-code-prompt-share.git
cd claude-code-prompt-share

# 依存関係をインストール
go mod download
```

## 使い方

```bash
# サーバーを起動
go run main.go
```

ブラウザで `http://localhost:8080` にアクセスしてください。

## プロジェクト構造

```
.
├── main.go              # アプリケーションのエントリーポイント
├── models/
│   └── models.go        # データモデル定義
├── services/
│   └── session_service.go  # セッションデータの読み込みロジック
├── handlers/
│   └── handlers.go      # HTTPハンドラー
├── templates/           # HTMLテンプレート
│   ├── index.html       # プロジェクト一覧
│   ├── project.html     # セッション一覧
│   ├── session.html     # 会話履歴
│   └── search.html      # 検索結果
└── static/
    └── style.css        # スタイルシート
```

## データ構造

Claude Codeは以下の構造でセッションデータを保存しています：

```
~/.claude/
├── projects/
│   └── [エンコードされたディレクトリパス]/
│       ├── [session-uuid].jsonl    # 会話履歴（JSONL形式）
│       └── [summary-uuid].jsonl    # 会話サマリー
├── settings.json                   # グローバル設定
└── commands/                       # カスタムコマンド
```

このアプリケーションは`.jsonl`ファイルを直接読み込み、パースして表示します。

## 技術スタック

- **Webフレームワーク**: [Echo](https://echo.labstack.com/) - 軽量で高性能なGoのWebフレームワーク
- **テンプレートエンジン**: Go標準の`html/template`
- **スタイリング**: カスタムCSS（グラデーション＆モダンなデザイン）

## 画面説明

### 左サイドバー（Projects & Sessions）
- プロジェクト一覧が表示されます
- プロジェクトをクリックすると、そのプロジェクトのセッション一覧が展開されます
- 検索ボックスでプロジェクトを絞り込めます

### 中央サイドバー（Prompts）
- 選択したセッションのユーザープロンプト（質問）が時系列で表示されます
- 各プロンプトには送信時刻とプレビューが表示されます

### 右メインエリア（Response）
- 選択したプロンプトと、それに対するAIの回答が表示されます
- プロンプトは紫色のグラデーション、回答は白い背景で表示されます
- タイムスタンプ付きで見やすく整理されています

## 開発

```bash
# 開発モードで起動（ホットリロード用）
go run main.go
```

## ライセンス

MIT

## 作者

Yugo Ibuki
