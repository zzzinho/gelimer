# LLM Reconciler 개선 작업

**작업 번호**: 001
**생성일**: 2025-10-09
**최종 수정일**: 2025-10-09 16:35 KST
**상태**: 진행중

## 작업 목록

- [ ] 작업 1: 헬스체크 Probe 설정 추가
- [ ] 작업 2: 에러 핸들링 및 재시도 로직 개선
- [ ] 작업 3: 롤링 업데이트 전략 구성
- [ ] 작업 4: GPU 리소스 설정 지원
- [ ] 작업 5: 상세한 Status 업데이트 로직 구현
- [ ] 작업 6: 리소스 상태 모니터링 및 검증
- [ ] 작업 7: Deployment 업데이트 전략 개선
- [ ] 작업 8: 리소스 정리 로직 개선

## 작업 상세 내용

### 작업 1: 헬스체크 Probe 설정 추가
**우선순위**: 높음
**의존성**: 없음
**대상 함수**: `buildVLLMDeployment()`

**작업 내용**:
- vLLM API 헬스 엔드포인트(/health, /v1/models)를 활용한 Liveness Probe 구성
- Readiness Probe 추가로 트래픽 수신 준비 상태 확인
- Startup Probe 추가로 모델 로딩 시간 고려
- Probe 설정을 LLM CRD Spec에서 선택적으로 재정의 가능하도록 구현

**완료 조건**:
- Container 스펙에 LivenessProbe, ReadinessProbe, StartupProbe 필드 추가
- 기본 Probe 설정이 vLLM의 특성에 맞게 구성됨 (예: initialDelaySeconds, periodSeconds)
- 대형 모델 로딩을 고려한 충분한 타임아웃 설정
- 사용자가 CRD Spec을 통해 Probe 설정을 오버라이드할 수 있음

**구현 위치**: `buildVLLMDeployment()` 함수 내 Container 정의 섹션

---

### 작업 2: 에러 핸들링 및 재시도 로직 개선
**우선순위**: 높음
**의존성**: 없음
**대상 함수**: `handleVLLM()`, `createOrUpdateVLLMResources()`

**작업 내용**:
- 일시적 오류와 영구적 오류를 구분하는 로직 구현
- 재시도 가능한 오류에 대한 Requeue 전략 개선
- 에러 타입별로 다른 Requeue 간격 설정 (예: 리소스 부족은 긴 간격, 일시적 API 오류는 짧은 간격)
- 에러 메시지를 Status에 명확히 기록
- 백오프 전략을 위한 ctrl.Result 활용

**완료 조건**:
- errors.IsNotFound, errors.IsConflict 등 에러 타입 별 처리 로직 구현
- RequeueAfter 필드를 활용한 재시도 간격 조정
- 영구적 오류 발생 시 명확한 에러 메시지와 함께 Status 업데이트
- 최대 재시도 횟수 추적 (Status에 RetryCount 필드 활용)

**구현 위치**:
- `handleVLLM()` 함수의 에러 처리 섹션
- `createOrUpdateVLLMResources()` 함수의 리턴 값 처리

---

### 작업 3: 롤링 업데이트 전략 구성
**우선순위**: 중간
**의존성**: 없음
**대상 함수**: `buildVLLMDeployment()`

**작업 내용**:
- Deployment.Spec.Strategy 필드 구성
- RollingUpdate 타입 설정 및 MaxUnavailable, MaxSurge 파라미터 정의
- LLM 특성을 고려한 안전한 업데이트 전략 (예: MaxUnavailable=0, MaxSurge=1로 무중단 배포)
- 사용자가 CRD Spec에서 업데이트 전략을 선택적으로 설정 가능

**완료 조건**:
- Deployment에 Strategy 필드가 명시적으로 설정됨
- 기본값으로 무중단 배포를 보장하는 전략 적용
- Pod Disruption Budget 고려사항 문서화 (별도 리소스이므로 구현은 제외)

**구현 위치**: `buildVLLMDeployment()` 함수의 DeploymentSpec 정의 섹션

---

### 작업 4: GPU 리소스 설정 지원
**우선순위**: 높음
**의존성**: 없음
**대상 함수**: `buildVLLMDeployment()`

**작업 내용**:
- LLM Spec의 Resources 필드에서 GPU 리소스 자동 감지
- nvidia.com/gpu 리소스 타입 지원
- GPU 사용 시 필요한 NodeSelector, Tolerations 자동 추가 옵션
- RuntimeClassName 설정 지원 (nvidia-container-runtime 등)
- 환경변수를 통한 GPU 관련 설정 전달 (CUDA_VISIBLE_DEVICES 등)

**완료 조건**:
- Resources에 nvidia.com/gpu가 명시된 경우 올바르게 반영됨
- GPU 노드 스케줄링을 위한 NodeSelector 설정 (선택적)
- RuntimeClassName 필드 추가 및 사용자 설정 가능
- GPU 관련 환경변수가 적절히 설정됨

**구현 위치**: `buildVLLMDeployment()` 함수의 PodSpec 정의 섹션

---

### 작업 5: 상세한 Status 업데이트 로직 구현
**우선순위**: 중간
**의존성**: 작업 6 (리소스 상태 모니터링)
**대상 함수**: `updateStatus()`, `createOrUpdateVLLMResources()`

**작업 내용**:
- Status에 Deployment의 실제 상태 반영 (Replicas, ReadyReplicas, UpdatedReplicas)
- Conditions 배열을 활용한 세부 상태 추적
- 각 단계별 상태 전이 명확히 구분 (Pending -> Creating -> Running -> Ready)
- 에러 발생 시 원인과 해결 방법 힌트 제공
- LastTransitionTime, ObservedGeneration 등 메타데이터 추가

**완료 조건**:
- Status.Conditions에 최소 3가지 타입 추가 (Available, Progressing, Degraded)
- Status.Replicas, Status.ReadyReplicas 필드 동기화
- Status.ObservedGeneration 업데이트
- 각 Condition에 Reason과 Message 필드가 명확히 설정됨

**구현 위치**:
- `updateStatus()` 함수 확장
- `createOrUpdateVLLMResources()` 함수에서 실제 리소스 상태 조회 후 Status 업데이트

---

### 작업 6: 리소스 상태 모니터링 및 검증
**우선순위**: 높음
**의존성**: 없음
**대상 함수**: `createOrUpdateVLLMResources()`

**작업 내용**:
- Deployment 생성/업데이트 후 실제 상태 확인
- Deployment.Status.Conditions 확인하여 배포 성공/실패 판단
- Pod 상태 조회 및 실패 원인 분석 (ImagePullBackOff, CrashLoopBackOff 등)
- Service 엔드포인트 생성 확인
- 리소스가 Ready 상태가 될 때까지 Requeue

**완료 조건**:
- Deployment 생성 후 Status.AvailableReplicas 확인
- Pod 상태가 Running이 아닌 경우 에러 원인 로깅
- Service의 Endpoints 생성 여부 확인
- 리소스가 준비되지 않은 경우 적절한 RequeueAfter 반환

**구현 위치**: `createOrUpdateVLLMResources()` 함수의 리소스 생성/업데이트 후 검증 로직

---

### 작업 7: Deployment 업데이트 전략 개선
**우선순위**: 중간
**의존성**: 없음
**대상 함수**: `createOrUpdateVLLMResources()`

**작업 내용**:
- 기존 Deployment 업데이트 시 불필요한 업데이트 방지
- Spec 변경 감지를 위한 해시 비교 로직 활용 강화
- Server-Side Apply 또는 Patch 방식으로 업데이트 최적화
- ResourceVersion 충돌 처리 개선
- 업데이트 시 롤백 가능성을 위한 RevisionHistoryLimit 설정

**완료 조건**:
- Deployment 업데이트 전 실제 변경사항 확인
- Patch 방식을 활용한 효율적인 업데이트
- Update 충돌 시 재시도 로직 구현
- RevisionHistoryLimit이 Deployment Spec에 설정됨

**구현 위치**: `createOrUpdateVLLMResources()` 함수의 Deployment 업데이트 섹션

---

### 작업 8: 리소스 정리 로직 개선
**우선순위**: 낮음
**의존성**: 없음
**대상 함수**: `cleanupResources()`

**작업 내용**:
- OwnerReference를 활용한 자동 삭제 확인
- 삭제 순서 최적화 (Service -> Deployment 순서로 연결 끊기 먼저)
- 삭제 완료 대기 로직 개선 (타임아웃 설정)
- 삭제 실패 시 상세한 에러 로깅
- Orphan 리소스 방지를 위한 검증

**완료 조건**:
- 리소스 삭제 시 타임아웃 설정 (예: 5분)
- 삭제가 완료될 때까지 Requeue하며 대기
- OwnerReference가 올바르게 설정되어 있는지 검증
- 삭제 실패 원인을 Status에 기록

**구현 위치**: `cleanupResources()` 함수

---

## 우선순위별 작업 순서

### Phase 1: 안정성 및 관측성 (높은 우선순위)
1. 작업 1: 헬스체크 Probe 설정 추가
2. 작업 2: 에러 핸들링 및 재시도 로직 개선
3. 작업 4: GPU 리소스 설정 지원
4. 작업 6: 리소스 상태 모니터링 및 검증

### Phase 2: 운영 효율성 (중간 우선순위)
5. 작업 3: 롤링 업데이트 전략 구성
6. 작업 5: 상세한 Status 업데이트 로직 구현
7. 작업 7: Deployment 업데이트 전략 개선

### Phase 3: 정리 및 최적화 (낮은 우선순위)
8. 작업 8: 리소스 정리 로직 개선

## 기술적 고려사항

### Reconciliation 패턴
- 모든 작업은 Reconciliation 루프의 멱등성(Idempotency)을 유지해야 함
- 상태 변경은 항상 선언적(Declarative) 방식으로 수행
- 외부 상태 조회 후 내부 상태를 동기화하는 패턴 준수

### 에러 처리 원칙
- 일시적 오류(Transient): Requeue with delay
- 영구적 오류(Permanent): Status 업데이트 후 중단
- 복구 가능한 오류: 재시도 횟수 제한과 함께 Requeue

### Status 업데이트 가이드라인
- Status는 항상 관측된 실제 상태를 반영
- Spec 변경과 무관하게 현재 리소스 상태를 정확히 기록
- Conditions를 통한 세부 상태 추적
- 사용자에게 유용한 메시지 제공

### GPU 지원 고려사항
- GPU 노드는 일반적으로 Taint가 설정되어 있음 (Toleration 필요)
- RuntimeClassName은 클러스터 구성에 따라 다를 수 있음
- GPU 리소스는 정수 단위로만 할당 가능

## 제외된 항목 및 사유

### HPA (Horizontal Pod Autoscaler)
- **사유**: Reconciler가 직접 HPA를 생성/관리하는 것은 책임 과다
- **대안**: 사용자가 LLM 리소스에 대해 별도로 HPA를 생성하도록 문서화

### Prometheus ServiceMonitor
- **사유**: 모니터링 인프라는 Reconciler의 관심사가 아님
- **대안**: Deployment에 적절한 Annotation 추가로 Prometheus 자동 발견 지원

### PVC 자동 생성 및 관리
- **사유**: 스토리지 관리는 별도의 관심사이며, 복잡도 증가
- **대안**: 사용자가 PVC를 미리 생성하고 Spec.Volumes에서 참조하도록 함

### 멀티 런타임 확장
- **사유**: reconciler.go 단일 파일에 모든 런타임 로직을 넣으면 유지보수 어려움
- **대안**: 향후 런타임별 별도 파일로 분리 (예: vllm.go, tgi.go)

### 테스트 코드
- **사유**: 테스트는 별도 파일(_test.go)로 관리
- **대안**: 별도 작업으로 단위 테스트 및 통합 테스트 추가

## 변경 이력

### 2025-10-09 16:35 KST
**변경 내용**: 파일명 수정 및 예상 소요 시간 필드 제거
**변경 이유**:
- 파일명 규칙: `{번호}_{요약}.md` 형식으로 통일 (날짜 형식 제거)
- 예상 소요 시간 제거: 시간 추정의 부정확성 방지 및 작업 유연성 향상
**영향받은 작업**: 전체 작업 (8개)

### 2025-10-09 16:32 KST
**변경 내용**: 최초 작업 계획 문서 생성
**변경 이유**: reconciler.go 파일의 안정성, 관측성, 운영 효율성 개선을 위한 체계적인 작업 계획 수립
**영향받은 작업**: 전체 작업 (8개)
