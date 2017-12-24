## goxz

Goのツールをパラレルクロスビルドして必要なファイルを抽出してarchiveに詰めてくれる君

- goxcから必要機能を抜き出して軽量にしたもの
- 古いGoのケアも不要
- デフォルトで良い感じになっているように

## usage

```console
% gozx -d ./dist \
    -build-ldflags=... -os=linux,darwin,windows -arch=amd64 \
    ./cmd/{{exename}} [...]

%  tree dist
dist/
├── {{Package}}_{{Version}}_{{GOOS}}_{{GOARCH}}.zip
└── ...
```

## 同梱物
- `LICENSE(?:.*)`
- `README(?:.*)`
- `INSTALL(?:.*)`
- `CREDIT(?:.*)`

## ファイル名の仕様
{{Package}}_{{Version}}_{{OS}}_{{Arch}}.zip

- Package(AppName?)はデフォルトでリポジトリ名
  - 複数実行ファイルを同梱するかもしれないので
- {{Version}} は入れるか否か
  - 入れるほうが好みだけど
  - 入れるのがデフォルトで、入れないオプションつくる？
    - バージョン指定ない場合は入れないとか(goxcもそうか)
- linuxのみ(BSD系も?)tar.gzにするのがデフォルトで、zipに統一するオプションを別途作るのが良い？
  - goxcはデフォではlinuxのみtar.gz
- テンプレート記法(最初はなくて良さそう)
  - goxc: {{.ExeName}}_{{.Version}}_{{.Os}}_{{.Arch}}{{.Ext}}
- 同梱物指定 (FileGatheringRule)
  - ゆくゆく

## その他オプション
- `os` and `arch`
   - osはdefault linux/darwin/win のみ(案)
   - archはdefault arm64 のみ(案)
- `-bc`
   - os/arch と bc それぞれは独立して動くで良い？
     - 良さそう
   - bc指定がされていて、 os/arch 設定が明にされていない場合は、デフォルトビルドはおこなわない
- `-pv` でバージョン指定
- `-build-ldflags` 実装は必須
- `-build-tags` はあっても良いかもなぁ
- `-o` オプションは作ってもよい？
  - あっても良いけど複数コマンドビルドとの食い合せが悪い
  - `go build` と同じ挙動で良さそうか (同じファイルを上書きする)

