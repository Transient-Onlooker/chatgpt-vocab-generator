# Go Vocabulary Generator

단어와 뜻을 입력하거나 텍스트 파일에서 불러와 영어 어휘 문제를 만드는 Go 터미널 앱입니다. Bubble Tea TUI에서 모델과 문제 유형을 고르고, OpenAI Chat Completions API로 생성을 요청합니다.

## 요구 사항

- `go.mod`에 지정된 Go 버전(1.25.4)
- OpenAI API 키

## 설정과 실행

저장소 루트에 `api.json`을 만들고 키를 설정합니다.

```json
{
  "chatgpt_api_key": "YOUR_OPENAI_API_KEY"
}
```

키 파일을 저장소에 커밋하지 마세요. 저장소 루트에서 실행합니다.

```powershell
go run ./src
```

TUI에서 `Ctrl+O`로 단어 텍스트 파일을 열고, `Ctrl+G`로 생성 흐름을 시작한 뒤 모델과 문제 유형을 선택합니다. `Ctrl+S`는 편집 내용을 파일로 저장합니다. 모델 선택 화면에는 GPT-5 pro, GPT-5, GPT-5 mini, GPT-5 nano, GPT-4.1이 정의되어 있습니다.

## 별도 Python GUI

`run_vocab_maker.bat`은 Go TUI가 아니라 `launcher.py`를 거쳐 `소스코드/gui_vocab_maker.py`를 실행합니다. 이 경로는 별도의 Python 의존성을 사용하므로, 실행 전에 `소스코드/requirements.txt`를 설치해야 합니다.

## 구조

- `src/main.go`: Go TUI 시작점
- `src/tui.go`: 화면과 키 입력 처리
- `src/parser.go`, `src/prompts.go`: 단어 입력 파싱과 문제 프롬프트
- `src/api.go`, `src/chatgpt.go`: API 키 로드와 OpenAI 요청
- `launcher.py`, `run_vocab_maker.bat`: Python GUI 실행 경로
