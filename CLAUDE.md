# CLAUDE.md

이 파일은 Claude Code (claude.ai/code)가 이 저장소의 코드로 작업할 때 지침을 제공합니다.

## 프로젝트 개요

Gelimer는 Kubernetes 환경에서 LLM(Large Language Model) 서빙과 AI 에이전트를 관리하는 Kubernetes Operator입니다. Kubebuilder를 사용하여 구축되었으며 Kubernetes 컨트롤러 패턴을 따릅니다.

## 주요 명령어

### 개발 명령어
- `make build` - 매니저 바이너리 빌드
- `make run` - 컨트롤러를 로컬에서 실행 (클러스터에 대한 kubectl 접근 필요)
- `make test` - 커버리지와 함께 단위 테스트 실행
- `make test-e2e` - Kind 클러스터를 사용하여 종단 간 테스트 실행
- `make fmt` - Go 코드 포맷팅
- `make vet` - go vet 정적 분석 실행
- `make lint` - golangci-lint 린터 실행
- `make lint-fix` - 자동 수정과 함께 린터 실행

### 코드 생성 명령어
- `make manifests` - CRD, RBAC, 웹훅 생성
- `make generate` - DeepCopy 메서드 및 기타 코드 생성

### Docker 및 배포 명령어
- `make docker-build` - 컨테이너 이미지 빌드
- `make docker-push` - 컨테이너 이미지 푸시
- `make install` - 클러스터에 CRD 설치
- `make deploy` - 클러스터에 컨트롤러 배포
- `make undeploy` - 클러스터에서 컨트롤러 제거

### 테스트 명령어
- `make setup-test-e2e` - e2e 테스트를 위한 Kind 클러스터 설정
- `make cleanup-test-e2e` - 테스트 클러스터 정리

## 아키텍처

### 핵심 컴포넌트

1. **메인 컨트롤러** (`internal/controller/controller.go`)
   - 특정 reconciler에 위임하는 진입점
   - LLM CRD 이벤트 처리

2. **LLM Reconciler** (`internal/controller/v1alpha1/llm/`)
   - `reconciler.go` - 메인 reconciliation 로직
   - `vllm.go` - vLLM 특화 구현
   - LLM 인스턴스를 위한 Deployment, Service, ConfigMap 관리

3. **API 타입** (`api/v1alpha1/`)
   - `llm_types.go` - LLM 커스텀 리소스 정의
   - 구성 가능한 매개변수로 vLLM 런타임 지원

4. **진입점** (`cmd/main.go`)
   - 매니저, 컨트롤러, 메트릭, 헬스 체크 설정
   - 웹훅과 메트릭을 위한 TLS 구성

### 주요 아키텍처 패턴

- **Operator 패턴**: reconciliation 루프를 위해 controller-runtime 사용
- **CRD 기반**: 커스텀 리소스를 통한 선언적 구성
- **모듈화된 Reconciler**: 리소스 타입별로 분리된 reconciler
- **런타임 추상화**: 다양한 LLM 런타임 지원 (현재 vLLM)

### 디렉터리 구조

```
├── api/v1alpha1/           # API 타입 정의 (CRD)
├── cmd/                    # 메인 애플리케이션 진입점
├── config/                 # Kubernetes 매니페스트 및 Kustomize 구성
├── internal/controller/    # 컨트롤러 구현
│   └── v1alpha1/llm/      # LLM 특화 reconciliation 로직
├── test/                   # 테스트 파일
├── hack/                   # 빌드 스크립트 및 유틸리티
└── bin/                    # 빌드된 바이너리
```

## 개발 워크플로우

1. **코드 생성**: API 타입 수정 후 항상 `make generate manifests` 실행
2. **테스트**: 변경사항 커밋 전 `make test` 실행
3. **린팅**: 코드는 `make lint` 검사를 통과해야 함
4. **로컬 개발**: 컨트롤러를 로컬에서 테스트하려면 `make run` 사용
5. **E2E 테스트**: 통합 테스트를 위해 `make test-e2e` 사용

## 구성 파일

- **Makefile**: 포괄적인 타겟을 가진 주요 빌드 시스템
- **.golangci.yml**: 엄격한 규칙을 가진 린터 구성
- **PROJECT**: Kubebuilder 프로젝트 메타데이터
- **go.mod**: Go 의존성 (Go 1.24.0 필요)

## 테스트

- 단위 테스트는 controller-runtime envtest 프레임워크 사용
- E2E 테스트는 격리를 위해 Kind 클러스터 사용
- 테스트는 KUBEBUILDER_ASSETS 환경 변수 필요
- 커버리지 리포트는 `cover.out`에 생성

## 중요 사항

- 프로젝트는 모듈 경로 `zzzinho.busan` 사용
- 리더 선출 ID: `22b90318.zzzinho.busan`
- 기본 컨테이너 이름: `model-container`
- vLLM 기본 포트: 8000
- 메트릭은 :8443 (HTTPS) 또는 :8080 (HTTP)에서 제공
- 헬스 프로브는 :8081에서 제공

## Claude 작업 규칙

### 언어 사용 규칙

Claude는 모든 상호작용에서 한국어를 우선적으로 사용해야 합니다.

#### 한국어 사용 원칙

1. **응답 언어**: 모든 응답과 설명은 한국어로 제공
2. **코드 주석**: Go 코드 내 주석은 한국어로 작성하여 이해도 향상
3. **문서화**: godoc 주석과 README 등 문서는 한국어로 작성
4. **에러 메시지**: 사용자 대상 에러 메시지는 한국어로 작성
5. **테스트 케이스**: 테스트 이름과 설명은 한국어로 작성하여 의도를 명확히 표현
6. **커밋 메시지**: Git 커밋 메시지는 한국어로 작성

#### 커밋 메시지 작성 규칙

**⚠️ 중요: 커밋 메시지에는 다음 문구들을 포함하지 않습니다:**
- "Generated with [Claude Code](https://claude.ai/code)"
- "Co-Authored-By: Claude <noreply@anthropic.com>"
- 기타 Claude 관련 자동 생성 문구

**커밋 메시지 형식**:
```
<type>(scope): <subject>

<body>
```

**예시**:
```
refactor(api): vLLM 설정을 args 기반으로 단순화

- VLLMConfig의 복잡한 중첩 구조체 제거
- Command와 Args 배열 기반의 단순한 설정 방식으로 변경
- vLLM 버전 의존성 문제 해결 및 유지보수성 향상
```

#### 예외 사항

1. **코드 식별자**: 변수명, 함수명, 타입명, 패키지명은 영어로 유지 (Go 언어 관례 준수)
2. **예약어**: Go 언어의 예약어와 표준 라이브러리는 원래 명칭 유지
3. **외부 API**: 외부 라이브러리나 API 호출은 원래 명칭 사용
4. **기술 용어**: 필요시 영어와 한국어를 병기 (예: "고루틴(goroutine)", "컨텍스트(context)")

#### 코드 예시

```go
// 사용자 인증을 처리하는 핸들러
func handleUserAuth(ctx context.Context, req *AuthRequest) (*AuthResponse, error) {
    if req.Username == "" {
        return nil, fmt.Errorf("사용자명이 필요합니다")
    }

    // 사용자 검증 로직
    user, err := validateUser(ctx, req.Username)
    if err != nil {
        return nil, fmt.Errorf("사용자 검증 실패: %w", err)
    }

    return &AuthResponse{User: user}, nil
}
```

### 작업 히스토리 관리

Claude는 모든 작업을 `.history/` 디렉터리에 기록해야 합니다.

#### 히스토리 파일 명명 규칙

**파일명 형식**: `YYYY-MM-DD_{number}_작업요약.md`

**번호 규칙**:
- `{number}`는 0부터 시작
- 같은 날짜에 히스토리가 추가될 때마다 1씩 증가
- 3자리 zero-padding 사용 (예: 000, 001, 002...)

**예시**:
- `2025-01-15_000_CLAUDE파일_한글번역.md` (첫 번째 작업)
- `2025-01-15_001_LLM_Reconciler_버그수정.md` (두 번째 작업)
- `2025-01-15_002_테스트_케이스_추가.md` (세 번째 작업)

#### 히스토리 파일 내용 구조

```markdown
# 작업 히스토리: [작업 요약]

**날짜**: YYYY-MM-DD HH:MM:SS
**작업자**: Claude

## 작업 개요
- 작업 내용에 대한 간단한 설명

## 수정된 파일
- `파일경로1` - 변경 사항 설명
- `파일경로2` - 변경 사항 설명

## 작업 상세 내용
상세한 작업 내용 설명

## 작업 목록
### 완료된 작업
- [x] 작업 1
- [x] 작업 2

### 미완료 작업
- [ ] 작업 3
- [ ] 작업 4

## 참고 사항
추가적인 참고 사항이나 향후 작업 방향
```

#### 작업 기록 의무사항

1. **모든 작업 기록**: 파일 수정, 생성, 삭제 등 모든 작업을 히스토리에 기록
2. **실시간 업데이트**: 작업 완료 즉시 히스토리 파일 생성 또는 업데이트
3. **명확한 요약**: 파일명과 내용에 작업 내용을 명확하게 기술
4. **작업 목록 관리**: TodoWrite 도구를 활용하여 진행 상황 추적

#### 히스토리 생성 강제 규칙

**⚠️ 중요: 다음 조건 중 하나라도 해당되면 반드시 히스토리 파일을 생성해야 합니다:**

1. **파일 변경**: 어떤 파일이든 수정, 생성, 삭제한 경우
2. **TodoWrite 완료**: TodoWrite에서 모든 작업이 completed 상태가 된 경우
3. **사용자 요청 완료**: 사용자가 요청한 작업이 완료된 경우
4. **복합 작업**: 여러 단계의 작업을 수행한 경우
5. **코드 생성/수정**: API 타입, 컨트롤러, 테스트 등 코드 관련 작업

#### 히스토리 생성 실행 절차

1. **작업 완료 후 즉시**: 마지막 응답 전에 반드시 히스토리 파일 생성
2. **파일명 생성**: 다음 단계를 따라 파일명 결정
   - 현재 날짜 확인: `YYYY-MM-DD` 형식
   - 같은 날짜의 기존 히스토리 파일 개수 확인
   - 번호 결정: 기존 개수를 기준으로 다음 번호 할당 (3자리 zero-padding)
   - 최종 파일명: `YYYY-MM-DD_{number}_작업요약.md`
3. **내용 작성**: 표준 템플릿을 사용하여 상세 내용 기록
4. **검증**: 히스토리 파일이 정상적으로 생성되었는지 확인

#### 히스토리 번호 생성 로직

**bash 명령어 예시**:
```bash
# 오늘 날짜 기준 기존 히스토리 파일 개수 확인
TODAY=$(date +"%Y-%m-%d")
COUNT=$(ls .history/${TODAY}_*.md 2>/dev/null | wc -l | tr -d ' ')
NUMBER=$(printf "%03d" $COUNT)
FILENAME="${TODAY}_${NUMBER}_작업요약.md"
```

#### 히스토리 누락 방지책

- **자동 알림**: 작업 완료 시 히스토리 생성 여부를 자동으로 점검
- **강제 실행**: 파일 변경이 감지되면 히스토리 생성을 강제로 요구
- **사용자 알림**: 히스토리가 누락된 경우 사용자에게 알림 및 즉시 생성
