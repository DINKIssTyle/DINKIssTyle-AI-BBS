// Created by DINKIssTyle on 2025. Copyright (C) 2025 DINKI'ssTyle. All rights reserved.

// Go 바인딩 함수들
const go = window.go?.main?.App || {};

// 상태 관리
const state = {
    isLoggedIn: false,
    currentUser: null,
    // 캐릭터 정렬 상태
    characterSort: { field: 'id', asc: true },
    cachedCharacters: [],
    cachedStats: {}
};

// DOM 요소 캐시
const $ = (selector) => document.querySelector(selector);
const $$ = (selector) => document.querySelectorAll(selector);

// 디바운스 유틸리티 - 연속 호출 시 마지막 호출만 실행
function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

// 캐릭터별 디바운스 함수 저장소
const characterUpdateDebounceMap = new Map();

// 초기화
document.addEventListener('DOMContentLoaded', async () => {
    // 앱 모드 확인
    try {
        const mode = await go.GetAppMode();
        console.log("App Mode:", mode);

        if (mode === 'char_manager') {
            initCharManagerMode();
        } else {
            initMainMode();
        }
    } catch (e) {
        console.error("Failed to get app mode:", e);
        initMainMode(); // 기본 fallback
    }
});

async function initMainMode() {
    const logContainer = $('#log-viewer-container');
    if (logContainer) logContainer.style.display = 'block';

    initEventListeners();
    await checkLoginStatus();
    await checkDBStatus();
    loadWebConfig();
    loadLLMConfig();
    loadBBSConfig();

    // 로그 이벤트 수신
    window.runtime?.EventsOn('log-event', (message) => {
        const viewer = $('#log-viewer');
        if (!viewer) return;

        const logItem = document.createElement('div');
        logItem.className = 'log-item';
        logItem.textContent = message.trim();
        viewer.appendChild(logItem);

        // 30줄 유지
        while (viewer.children.length > 30) {
            viewer.removeChild(viewer.firstChild);
        }

        // 자동 스크롤
        viewer.scrollTop = viewer.scrollHeight;
    });

    // 새 창 열기 버튼 이벤트
    const btnOpen = $('#btn-open-char-manager');
    if (btnOpen) {
        btnOpen.addEventListener('click', async () => {
            try {
                await go.OpenCharacterManagerWindow();
            } catch (e) {
                alert("새 창 열기 실패: " + e);
            }
        });
    }
}

async function initCharManagerMode() {
    // DOM 정리: 중복 ID 방지를 위해 사용하지 않는 메인 UI 요소들 제거
    $('#header')?.remove();
    $('.settings-tabs')?.remove();
    $$('.tab-content').forEach(el => el.remove()); // 기존 탭들(tab-ai 포함) 제거

    // 대시보드 뷰 스타일 정리 (부모 컨테이너) -> Flex item으로 변경되어야 함
    const dashboard = $('#dashboard-view');
    if (dashboard) {
        dashboard.style.padding = '0';
        dashboard.style.margin = '0';
        dashboard.style.height = '100%'; // Full height
        dashboard.style.flex = '1';
        dashboard.style.overflow = 'hidden'; // 내부 스크롤 방지 (Manager view가 함)
        dashboard.style.display = 'flex';
        dashboard.style.flexDirection = 'column';
    }

    // 캐릭터 관리자 전용 뷰 설정 (Flexbox 풀스크린)
    const managerView = $('#view-char-manager-window');
    if (managerView) {
        managerView.style.display = 'flex';
        managerView.style.flexDirection = 'column';
        managerView.style.flex = '1'; // 남은 공간 차지
        managerView.style.width = '100%';
        managerView.style.height = '100%';
        managerView.style.minHeight = '0'; // Flex item shrinking fix
        managerView.style.padding = '20px';
        managerView.style.background = 'var(--bg-color)';
        managerView.style.boxSizing = 'border-box';
        // Position fixed 제거 -> App Layout(Footer)을 따름
    }

    // 테이블 컨테이너 설정 (남은 공간 모두 차지 + 스크롤)
    const container = $('#character-table-container');
    if (container) {
        container.style.height = 'auto'; // 기존 인라인 스타일 무시
        container.style.flex = '1';
        container.style.overflowY = 'auto';
        container.style.minHeight = '0'; // Flex item scroll fix
        container.style.border = '1px solid var(--border-color)';
    }

    // 필수 이벤트 리스너만 바인딩
    const btnGen = $('#btn-generate-chars');
    if (btnGen) btnGen.addEventListener('click', generateCharacters);

    const tbody = $('#character-tbody');
    if (tbody) {
        // 캐릭터 테이블 이벤트 위임 (삭제 등)
        tbody.addEventListener('change', (e) => {
            const tr = e.target.closest('tr');
            if (tr && (e.target.matches('input') || e.target.matches('select'))) {
                const btn = tr.querySelector('.btn-delete-char');
                if (btn) {
                    updateCharacter(parseInt(btn.dataset.id), tr);
                }
            }
        });

        tbody.addEventListener('click', (e) => {
            if (e.target.matches('.btn-delete-char')) {
                const id = parseInt(e.target.dataset.id);
                deleteCharacter(id);
            }
        });
    }

    // 초기 데이터 로드
    loadCharacters();

    // 백엔드 로그 수신
    if (window.runtime) {
        window.runtime.EventsOn("debug_log", (msg) => {
            console.log(msg);
        });
    }
}

// 이벤트 리스너 초기화
function initEventListeners() {
    // 탭 전환
    $$('.tab-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            $$('.tab-btn').forEach(b => b.classList.remove('active'));
            $$('.tab-content').forEach(c => c.classList.remove('active'));
            e.target.classList.add('active');

            const tabId = e.target.dataset.tab;
            $(`#tab-${tabId}`).classList.add('active');

            if (tabId === 'account') loadUsers();
            if (tabId === 'ai') loadCharacters();
            if (tabId === 'server') loadWebConfig();
            if (tabId === 'bbs') loadBBSConfig();
            if (tabId === 'database') {
                loadDatabaseList();
                loadDatabaseInfo();
            }
            if (tabId === 'ai-prompts') {
                loadCompPrompts();
                loadMBTIList();
            }
        });
    });

    // 계정 관리
    $('#btn-do-login').addEventListener('click', doLogin);
    $('#btn-do-logout').addEventListener('click', doLogout);
    $('#btn-show-profile-modal').addEventListener('click', () => showModal('profile-modal'));
    $('#btn-show-create-user-modal').addEventListener('click', () => showModal('create-user-modal'));
    $('#btn-create-user-submit').addEventListener('click', registerUserInModal);

    $('#btn-change-password').addEventListener('click', changePassword);
    $('#btn-change-nickname').addEventListener('click', changeNickname);

    // 데이터베이스 관리
    $('#btn-reset-db').addEventListener('click', () => showModal('confirm-modal'));
    $('#btn-confirm-reset').addEventListener('click', resetDatabase);
    $('#btn-cancel-reset').addEventListener('click', () => hideModal('confirm-modal'));
    $('#btn-create-db')?.addEventListener('click', createNewDatabase);
    $('#db-select')?.addEventListener('change', switchDatabase);
    $('#btn-delete-db')?.addEventListener('click', () => showModal('delete-modal'));
    $('#btn-confirm-delete')?.addEventListener('click', deleteDatabase);
    $('#btn-cancel-delete')?.addEventListener('click', () => hideModal('delete-modal'));

    // LLM 관리
    $('#btn-test-llm').addEventListener('click', testLLMConnection);
    $('#btn-save-llm').addEventListener('click', saveLLMConfig);
    $('#btn-start-ai').addEventListener('click', startAIActivity);
    $('#btn-stop-ai').addEventListener('click', stopAIActivity);

    // 모델 입력 지우기 버튼
    $('#btn-clear-model-1').addEventListener('click', () => $('#llm-model-1').value = '');
    $('#btn-clear-model-2').addEventListener('click', () => $('#llm-model-2').value = '');
    $('#btn-clear-model-3').addEventListener('click', () => $('#llm-model-3').value = '');

    // AI 캐릭터 관리
    $('#btn-generate-chars').addEventListener('click', generateCharacters);

    // 웹 서버 관리
    $('#btn-start-web').addEventListener('click', startWebServer);
    $('#btn-stop-web').addEventListener('click', stopWebServer);

    // 웹 설정 저장
    $('#btn-save-web-config')?.addEventListener('click', saveWebConfig);

    // SSL 설정 표시 토글
    $('#web-ssl-enabled').addEventListener('change', (e) => {
        const group = $('#ssl-settings-group');
        if (group) {
            group.style.display = e.target.value === 'true' ? 'block' : 'none';
        }
    });

    // SSL 파일 선택 (찾아보기)
    $('#btn-browse-cert')?.addEventListener('click', async () => {
        try {
            const path = await go.SelectFile('SSL 인증서 선택', '인증서 파일 (*.crt, *.pem)');
            if (path) $('#web-ssl-cert').value = path;
        } catch (e) { console.error(e); }
    });

    $('#btn-browse-key')?.addEventListener('click', async () => {
        try {
            const path = await go.SelectFile('SSL 비밀키 선택', '비밀키 파일 (*.key, *.pem)');
            if (path) $('#web-ssl-key').value = path;
        } catch (e) { console.error(e); }
    });

    // BBS 설정
    $('#btn-save-bbs').addEventListener('click', saveBBSConfig);

    // 모달 닫기
    $$('.modal-close').forEach(btn => {
        btn.addEventListener('click', () => {
            btn.closest('.modal').classList.remove('active');
        });
    });

    // 이벤트 위임
    $('#user-tbody').addEventListener('change', async (e) => {
        if (e.target.matches('.admin-check')) {
            const checkbox = e.target;
            const userId = parseInt(checkbox.dataset.id);
            const newStatus = checkbox.checked;

            // 관리자 권한 체크 제거 (요청사항 반영)
            // if (!state.currentUser || !state.currentUser.is_admin) { ... }

            try {
                await go.SetUserAdmin(userId, newStatus);
                showToast(`관리자 권한을 ${newStatus ? '부여' : '해제'}했습니다.`);
            } catch (err) {
                checkbox.checked = !newStatus;
                showToast('권한 변경 실패: ' + err, 'error');
            }
        }
    });

    $('#character-tbody').addEventListener('change', (e) => {
        const tr = e.target.closest('tr');
        if (tr && (e.target.matches('input') || e.target.matches('select'))) {
            const btn = tr.querySelector('.btn-delete-char');
            if (btn) {
                updateCharacter(parseInt(btn.dataset.id), tr);
            }
        }
    });

    $('#character-tbody').addEventListener('click', (e) => {
        if (e.target.matches('.btn-delete-char')) {
            const id = parseInt(e.target.dataset.id);
            deleteCharacter(id);
        }
    });

    // AI 프롬프트 관리 이벤트 위임
    $('#tab-ai-prompts').addEventListener('click', (e) => {
        if (e.target.matches('.btn-save-prompt')) {
            saveCompPrompt(e.target.dataset.key);
        }
        if (e.target.matches('.btn-reset-prompt')) {
            resetCompPrompt(e.target.dataset.key);
        }
        if (e.target.matches('.btn-save-mbti')) {
            const mbti = e.target.dataset.mbti;
            const input = e.target.closest('tr').querySelector('.input-mbti-desc');
            saveMBTI(mbti, input.value);
        }
        if (e.target.matches('.btn-reset-mbti')) {
            resetMBTI(e.target.dataset.mbti);
        }
    });
}

// 모달 표시/숨김
function showModal(id) {
    if ($(`#${id}`)) $(`#${id}`).classList.add('active');
}

function hideModal(id) {
    if ($(`#${id}`)) $(`#${id}`).classList.remove('active');
}

// 토스트 메시지
function showToast(message, type = 'info') {
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    toast.textContent = message;
    document.body.appendChild(toast);
    setTimeout(() => toast.remove(), 3000);
}

// 유틸리티
function formatDate(dateStr) {
    const date = new Date(dateStr);
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    return `${month}-${day}`;
}

// HTML 이스케이프 함수
function escapeHtml(text) {
    if (!text) return '';
    return text
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#039;");
}

// 계정 함수
async function checkLoginStatus() {
    try {
        const user = await go.GetCurrentUser();
        if (user) {
            state.isLoggedIn = true;
            state.currentUser = user;
            updateLoginUI();
        }
    } catch (e) {
        console.log('Not logged in');
    }
}

function updateLoginUI() {
    if (state.isLoggedIn && state.currentUser) {
        $('#account-guest').style.display = 'none';
        $('#account-logged').style.display = 'block';
        $('#account-nickname').textContent = state.currentUser.nickname;

        if (state.currentUser.is_admin) {
            if ($('#admin-badge')) $('#admin-badge').style.display = 'inline';
        } else {
            if ($('#admin-badge')) $('#admin-badge').style.display = 'none';
        }
    } else {
        $('#account-guest').style.display = 'block';
        $('#account-logged').style.display = 'none';
    }
}

async function doLogin() {
    const username = $('#login-username').value.trim();
    const password = $('#login-password').value;
    try {
        const user = await go.Login(username, password);
        state.isLoggedIn = true;
        state.currentUser = user;
        updateLoginUI();
        showToast(`환영합니다, ${user.nickname}님!`);
        $('#login-username').value = '';
        $('#login-password').value = '';
        loadUsers();
    } catch (e) { showToast('로그인 실패: ' + e, 'error'); }
}

async function doLogout() {
    try {
        await go.Logout();
        state.isLoggedIn = false;
        state.currentUser = null;
        updateLoginUI();
        showToast('로그아웃되었습니다');
        loadUsers();
    } catch (e) { showToast('로그아웃 실패: ' + e, 'error'); }
}

async function registerUserInModal() {
    const username = $('#create-username').value.trim();
    const password = $('#create-password').value;
    const nickname = $('#create-nickname').value.trim();
    if (!username || !password || !nickname) {
        showToast('모든 필드를 입력해주세요', 'error'); return;
    }
    try {
        await go.CreateUser(username, password, nickname);
        showToast('계정이 생성되었습니다.');
        $('#create-username').value = ''; $('#create-password').value = ''; $('#create-nickname').value = '';
        hideModal('create-user-modal');
        loadUsers();
    } catch (e) { showToast('계정 생성 실패: ' + e, 'error'); }
}

async function changePassword() {
    const oldPw = $('#old-password').value;
    const newPw = $('#new-password').value;
    try {
        await go.ChangePassword(oldPw, newPw);
        showToast('비밀번호가 변경되었습니다');
        $('#old-password').value = ''; $('#new-password').value = '';
        hideModal('profile-modal');
    } catch (e) { showToast('비밀번호 변경 실패: ' + e, 'error'); }
}

async function changeNickname() {
    const newNickname = $('#new-nickname').value.trim();
    try {
        await go.ChangeNickname(newNickname);
        state.currentUser.nickname = newNickname;
        updateLoginUI();
        showToast('닉네임이 변경되었습니다');
        $('#new-nickname').value = '';
        hideModal('profile-modal');
    } catch (e) { showToast('닉네임 변경 실패: ' + e, 'error'); }
}

async function loadUsers() {
    try {
        const users = await go.GetAllUsers();
        renderUsers(users);
    } catch (e) { console.log('유저 목록 로딩 실패:', e); }
}

function renderUsers(users) {
    const tbody = $('#user-tbody');
    tbody.innerHTML = '';
    if (!users || users.length === 0) {
        tbody.innerHTML = '<tr><td colspan="5" style="text-align:center;">회원이 없습니다.</td></tr>';
        return;
    }

    // const isAdmin = state.currentUser && state.currentUser.is_admin;

    users.forEach(user => {
        const tr = document.createElement('tr');
        // const disabledAttr = isAdmin ? '' : 'disabled';

        tr.innerHTML = `
            <td>${user.id}</td>
            <td>${escapeHtml(user.username)}</td>
            <td>${escapeHtml(user.nickname)}</td>
            <td>
                <input type="checkbox" class="admin-check" data-id="${user.id}" 
                    ${user.is_admin ? 'checked' : ''} />
            </td>
            <td>${formatDate(user.created_at)}</td>
        `;
        tbody.appendChild(tr);
    });
}

// 서버 및 DB, AI 함수
async function checkDBStatus() {
    try {
        const connected = await go.IsDatabaseConnected();
        if ($('#db-status-text')) $('#db-status-text').textContent = connected ? '연결됨' : '연결 안됨';
        if (connected) {
            const aiRunning = await go.IsAIActivityRunning();
            updateAIStatus(aiRunning);
            // DB 목록 및 정보 로드
            loadDatabaseList();
            loadDatabaseInfo();
        }
    } catch (e) { console.log('DB 상태 확인 실패:', e); }
}

// DB 목록 로드
async function loadDatabaseList() {
    try {
        const list = await go.GetDatabaseList();
        const current = await go.GetCurrentDatabase();
        const select = $('#db-select');
        if (!select) return;

        select.innerHTML = '';
        list.forEach(name => {
            const opt = document.createElement('option');
            opt.value = name;
            opt.textContent = name;
            if (name === current) opt.selected = true;
            select.appendChild(opt);
        });
    } catch (e) {
        console.log('DB 목록 로드 실패:', e);
    }
}

// DB 정보 로드
async function loadDatabaseInfo() {
    try {
        const info = await go.GetDatabaseInfo();
        if ($('#db-info-title')) $('#db-info-title').textContent = info.title || '-';
        if ($('#db-info-system-role')) {
            const role = info.systemRole || '-';
            $('#db-info-system-role').textContent = role.length > 50 ? role.substring(0, 50) + '...' : role;
        }
        if ($('#db-info-posts')) $('#db-info-posts').textContent = info.postCount || 0;
        if ($('#db-info-comments')) $('#db-info-comments').textContent = info.commentCount || 0;
        if ($('#db-info-size')) $('#db-info-size').textContent = formatFileSize(info.size || 0);

        // 헤더에 현재 DB 정보 표시
        if ($('#header-db-info')) {
            const title = info.title || 'DINKIssTyle AI BBS';
            const dbName = info.name || 'default.db';
            $('#header-db-info').textContent = `${title} - ${dbName}`;
        }
    } catch (e) {
        console.log('DB 정보 로드 실패:', e);
    }
}

// 파일 크기 포맷
function formatFileSize(bytes) {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

// DB 전환
async function switchDatabase() {
    const select = $('#db-select');
    if (!select) return;

    const name = select.value;
    if (!name) return;

    const current = await go.GetCurrentDatabase();
    if (name === current) return;

    if (!confirm(`"${name}" 데이터베이스로 전환하시겠습니까?\n\n※ 웹 서버와 AI 활동이 중지됩니다.`)) {
        // 원래 선택으로 복원
        loadDatabaseList();
        return;
    }

    try {
        await go.SwitchDatabase(name);
        showToast(`"${name}"로 전환되었습니다`);

        // 웹서버/AI 상태 인디케이터 업데이트 (정지됨)
        updateWebStatus(false);
        updateAIStatus(false);

        loadDatabaseInfo();
        loadDatabaseList();

        // 웹서버 재시작 확인
        if (confirm('웹 서버를 다시 시작하시겠습니까?')) {
            await startWebServer();
        }
    } catch (e) {
        showToast('DB 전환 실패: ' + e, 'error');
        loadDatabaseList();
    }
}

// 새 DB 생성
async function createNewDatabase() {
    const name = prompt('새 데이터베이스 파일 이름을 입력하세요.\n(확장자 .db는 자동 추가됩니다)');
    if (!name) return;

    // 유효성 검사
    const safeName = name.replace(/[^a-zA-Z0-9_-]/g, '');
    if (safeName !== name) {
        showToast('파일 이름에는 영문, 숫자, _, - 만 사용 가능합니다', 'error');
        return;
    }

    try {
        await go.CreateNewDatabase(safeName);
        showToast(`새 데이터베이스 "${safeName}.db"가 생성되었습니다`);
        loadDatabaseList();
        loadDatabaseInfo();
    } catch (e) {
        showToast('DB 생성 실패: ' + e, 'error');
    }
}

async function resetDatabase() {
    const confirmation = $('#confirm-input').value;
    try {
        await go.ResetDatabase(confirmation);
        showToast('데이터베이스가 초기화되었습니다');
        hideModal('confirm-modal');
        $('#confirm-input').value = '';
        loadUsers();
        loadDatabaseInfo();
        loadDatabaseList();
        const chars = await go.GetAllCharacters();
        renderCharacters(chars);
    } catch (e) { showToast('초기화 실패: ' + e, 'error'); }
}

async function deleteDatabase() {
    const confirmation = $('#delete-confirm-input').value;
    try {
        await go.DeleteDatabase(confirmation);
        showToast('데이터베이스가 삭제되었습니다');
        hideModal('delete-modal');
        $('#delete-confirm-input').value = '';

        // 상태 업데이트
        updateWebStatus(false);
        updateAIStatus(false);

        loadDatabaseList();
        loadDatabaseInfo();
        loadUsers();
    } catch (e) { showToast('삭제 실패: ' + e, 'error'); }
}

function updateAIStatus(running) {
    const headerDot = $('#header-ai-dot');
    if ($('#ai-activity-status')) {
        $('#ai-activity-status').textContent = running ? '실행중' : '정지됨';
        $('#ai-activity-status').className = running ? 'value status-on' : 'value status-off';
    }
    if (headerDot) {
        headerDot.className = running ? 'status-dot status-on' : 'status-dot status-off';
    }
}

async function startAIActivity() {
    try {
        await go.StartAIActivity();
        updateAIStatus(true);
        showToast('AI 활동이 시작되었습니다');
    } catch (e) { showToast('AI 활동 시작 실패: ' + e, 'error'); }
}

async function stopAIActivity() {
    try {
        await go.StopAIActivity();
        updateAIStatus(false);
        showToast('AI 활동이 정지되었습니다');
    } catch (e) { showToast('AI 활동 정지 실패: ' + e, 'error'); }
}

async function loadLLMConfig() {
    try {
        const config = await go.GetLLMConfig();
        if (config) {
            $('#llm-host').value = config.host;
            $('#llm-port').value = config.port;

            $('#llm-model-1').value = config.model1 || 'default';
            $('#llm-model-2').value = config.model2 || '';
            $('#llm-model-3').value = config.model3 || '';

            $('#llm-posts').value = config.posts_per_hour;
            $('#llm-comments').value = config.comments_per_hour;
            $('#llm-tokens').value = config.max_tokens;
            $('#llm-temp').value = config.temperature;
        }
    } catch (e) { console.log('LLM 설정 로드 실패', e); }
}

async function saveLLMConfig() {
    const host = $('#llm-host').value;
    const port = $('#llm-port').value;
    const model1 = $('#llm-model-1').value;
    const model2 = $('#llm-model-2').value;
    const model3 = $('#llm-model-3').value;

    const postsPerHour = parseInt($('#llm-posts').value) || 5;
    const commentsPerHour = parseInt($('#llm-comments').value) || 10;
    const maxTokens = parseInt($('#llm-tokens').value) || 2000;
    const temperature = parseFloat($('#llm-temp').value) || 0.8;
    const timeout = parseInt($('#llm-timeout').value) || 120;

    try {
        await go.SaveLLMConfig(host, port, model1, model2, model3, postsPerHour, commentsPerHour, maxTokens, temperature, timeout);
        showToast('LLM 설정이 저장되었습니다');
    } catch (e) { showToast('설정 저장 실패: ' + e, 'error'); }
}

async function testLLMConnection() {
    const host = $('#llm-host').value;
    const port = $('#llm-port').value;
    let model = $('#llm-model-1').value;
    if (!model) model = $('#llm-model-2').value;
    if (!model) model = $('#llm-model-3').value;
    if (!model) model = 'default';

    try {
        await go.TestLLMConnection(host, port, model);
        showToast(`LLM 연결 성공! (모델: ${model})`);
    } catch (e) { showToast('LLM 연결 실패: ' + e, 'error'); }
}

async function generateCharacters() {
    const count = parseInt($('#char-count').value) || 10;
    try {
        const chars = await go.GenerateCharacters(count);
        showToast(`${chars.length}명의 캐릭터가 생성되었습니다`);
        loadCharacters();
    } catch (e) { showToast('캐릭터 생성 실패: ' + e, 'error'); }
}

async function loadCharacters() {
    try {
        const chars = await go.GetAllCharacters();
        const stats = await go.GetAllCharacterStats();
        // 아바타 및 MBTI 캐싱
        state.avatarCache = {
            male: await go.GetAvatarList("male"),
            female: await go.GetAvatarList("female")
        };
        state.mbtiTypes = await go.GetMBTITypes();
        state.cachedCharacters = chars || [];
        state.cachedStats = stats || {};
        sortAndRenderCharacters();
    } catch (e) { console.log('캐릭터 로딩 실패:', e); }
}

function sortAndRenderCharacters() {
    const { field, asc } = state.characterSort;
    const chars = [...(state.cachedCharacters || [])];
    const stats = state.cachedStats || {};

    chars.sort((a, b) => {
        let valA, valB;

        // 특수 필드 처리 (글/댓글 수)
        if (field === 'post_count') {
            valA = (stats[a.id] || {}).post_count || 0;
            valB = (stats[b.id] || {}).post_count || 0;
        } else if (field === 'comment_count') {
            valA = (stats[a.id] || {}).comment_count || 0;
            valB = (stats[b.id] || {}).comment_count || 0;
        } else {
            valA = a[field];
            valB = b[field];
        }

        // null/undefined 처리
        if (valA == null) valA = '';
        if (valB == null) valB = '';

        // 숫자 비교
        if (typeof valA === 'number' && typeof valB === 'number') {
            return asc ? valA - valB : valB - valA;
        }

        // 문자열 비교
        const strA = String(valA).toLowerCase();
        const strB = String(valB).toLowerCase();
        if (strA < strB) return asc ? -1 : 1;
        if (strA > strB) return asc ? 1 : -1;
        return 0;
    });

    // 검색 필터 적용
    const searchInput = $('#char-search');
    const searchTerm = searchInput ? searchInput.value : '';
    const filteredChars = filterCharactersBySearch(chars, searchTerm);

    renderCharacters(filteredChars, stats);
    updateSortIndicators();
}

function handleCharacterSort(field) {
    if (state.characterSort.field === field) {
        state.characterSort.asc = !state.characterSort.asc;
    } else {
        state.characterSort.field = field;
        state.characterSort.asc = true;
    }
    sortAndRenderCharacters();
}

function updateSortIndicators() {
    const headers = $$('#character-table th');
    headers.forEach(th => {
        th.classList.remove('sort-asc', 'sort-desc');
        const sortField = th.dataset.sort;
        if (sortField === state.characterSort.field) {
            th.classList.add(state.characterSort.asc ? 'sort-asc' : 'sort-desc');
        }
    });
}

// 캐릭터 검색 필터링
function filterCharactersBySearch(chars, searchTerm) {
    if (!searchTerm || searchTerm.trim() === '') {
        return chars;
    }
    const term = searchTerm.toLowerCase().trim();
    return chars.filter(c => {
        const searchFields = [
            c.id?.toString() || '',
            c.nickname || '',
            c.gender || '',
            c.age?.toString() || '',
            c.birthdate || '',
            c.region || '',
            c.job_category || '',
            c.hobby || '',
            c.mbti || '',
            c.aggression_level?.toString() || '',
            c.formality_level?.toString() || '',
            c.persona_summary || '',
            c.post_count?.toString() || '',
            c.comment_count?.toString() || '',
            c.assigned_model_index?.toString() || ''
        ];
        return searchFields.some(field => field.toLowerCase().includes(term));
    });
}

// 검색창 이벤트 리스너
const charSearchInput = $('#char-search');
if (charSearchInput) {
    charSearchInput.addEventListener('input', () => {
        sortAndRenderCharacters();
    });
}

function renderCharacters(chars, stats = {}) {
    const tbody = $('#character-tbody');
    tbody.innerHTML = '';
    if (!chars || chars.length === 0) {
        tbody.innerHTML = '<tr><td colspan="17" style="text-align:center;padding:20px;">캐릭터가 없습니다.</td></tr>';
        return;
    }
    chars.forEach(c => {
        const charStats = stats[c.id] || { post_count: 0, comment_count: 0 };
        const tr = document.createElement('tr');
        tr.dataset.charId = c.id;
        tr.innerHTML = `
            <td><input type="checkbox" class="char-select-checkbox" data-char-id="${c.id}"/></td>
            <td>${c.id}</td>
            <td><input type="text" value="${escapeHtml(c.nickname)}" data-field="nickname" style="width:80px;"/></td>
            <td>
                <select data-field="gender" onchange="handleTableChange(this, 'gender', ${c.id})">
                    <option value="남성" ${c.gender === '남성' ? 'selected' : ''}>남성</option>
                    <option value="여성" ${c.gender === '여성' ? 'selected' : ''}>여성</option>
                </select>
            </td>
            <td>
                <div style="display:flex; align-items:center; gap:5px;">
                    <img src="/avarta/${c.gender === '남성' ? 'male' : 'female'}/${c.avatar_image}" 
                         style="width:30px; height:30px; border-radius:50%; object-fit:cover;" 
                         onerror="this.src='data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7'"
                         id="avatar-preview-${c.id}">
                    <select data-field="avatar_image" style="width:105px;" onchange="handleTableChange(this, 'avatar_image', ${c.id})">
                        <option value="">(없음)</option>
                        ${(state.avatarCache[c.gender === '남성' ? 'male' : 'female'] || []).map(img =>
            `<option value="${img}" ${c.avatar_image === img ? 'selected' : ''}>${img}</option>`
        ).join('')}
                    </select>
                </div>
            </td>
            <td><input type="number" value="${c.age}" min="10" max="80" data-field="age" style="width:40px;"/></td>
            <td><input type="text" value="${c.birthdate || ''}" data-field="birthdate" placeholder="YYYY-MM-DD" style="width:90px;"/></td>
            <td><input type="text" value="${escapeHtml(c.region) || ''}" data-field="region" style="width:60px;"/></td>
            <td><input type="text" value="${escapeHtml(c.job_category)}" data-field="job_category" style="width:100px;"/></td>
            <td><input type="text" value="${escapeHtml(c.hobby) || ''}" data-field="hobby" style="width:80px;"/></td>
            <td>
                <select data-field="mbti" style="width:70px;">
                    ${state.mbtiTypes.map(mbti => `<option value="${mbti}" ${c.mbti === mbti ? 'selected' : ''}>${mbti}</option>`).join('')}
                </select>
            </td>
            <td><input type="number" value="${c.aggression_level}" min="1" max="10" data-field="aggression_level" style="width:40px;"/></td>
            <td><input type="number" value="${c.formality_level}" min="1" max="10" data-field="formality_level" style="width:40px;"/></td>
            <td><input type="text" value="${escapeHtml(c.persona_summary) || ''}" data-field="persona_summary" placeholder="인격 요약" style="width:150px;"/></td>
            <td style="text-align:center;">${charStats.post_count}</td>
            <td style="text-align:center;">${charStats.comment_count}</td>
            <td>
                <span class="badge">${c.assigned_model_index || 1}</span>
            </td>
            <td>
                <input type="checkbox" ${c.is_active ? 'checked' : ''} data-field="is_active"/>
            </td>
            <td>
                <button class="btn-delete-char" data-id="${c.id}">삭제</button>
            </td>
            
            <input type="hidden" data-field="roleplay_level" value="${c.roleplay_level}" />
            <input type="hidden" data-field="assigned_model_index" value="${c.assigned_model_index}" />
        `;
        tbody.appendChild(tr);
    });
}

async function updateCharacter(id, tr) {
    if (!tr) return;
    const data = {
        id: parseInt(id),
        nickname: tr.querySelector('[data-field="nickname"]').value,
        gender: tr.querySelector('[data-field="gender"]').value,
        avatar_image: tr.querySelector('[data-field="avatar_image"]').value,
        age: parseInt(tr.querySelector('[data-field="age"]').value),
        birthdate: tr.querySelector('[data-field="birthdate"]').value,
        region: tr.querySelector('[data-field="region"]').value,
        job_category: tr.querySelector('[data-field="job_category"]').value,
        hobby: tr.querySelector('[data-field="hobby"]').value,
        mbti: tr.querySelector('[data-field="mbti"]').value,

        // Levels
        aggression_level: parseInt(tr.querySelector('[data-field="aggression_level"]').value),
        formality_level: parseInt(tr.querySelector('[data-field="formality_level"]').value),
        roleplay_level: parseInt(tr.querySelector('[data-field="roleplay_level"]').value),
        persona_summary: tr.querySelector('[data-field="persona_summary"]').value,
        assigned_model_index: parseInt(tr.querySelector('[data-field="assigned_model_index"]').value),

        is_active: tr.querySelector('[data-field="is_active"]').checked,
    };
    try { await go.UpdateCharacter(data); }
    catch (e) { showToast('캐릭터 수정 실패: ' + e, 'error'); }
}

// 디바운스된 캐릭터 업데이트 (캐릭터 ID별로 별도 디바운스)
function debouncedUpdateCharacter(id, tr) {
    if (!characterUpdateDebounceMap.has(id)) {
        characterUpdateDebounceMap.set(id, debounce((charId, charTr) => {
            updateCharacter(charId, charTr);
        }, 500));
    }
    characterUpdateDebounceMap.get(id)(id, tr);
}

async function deleteCharacter(id) {
    if (!confirm('정말 삭제하시겠습니까?')) return;
    try {
        await go.DeleteCharacter(id);
        showToast('캐릭터가 삭제되었습니다');
        loadCharacters();
    } catch (e) { showToast('캐릭터 삭제 실패: ' + e, 'error'); }
}

async function startWebServer() {
    const port = $('#web-port').value;
    const registration = $('#web-registration').value === 'true';
    const sslEnabled = $('#web-ssl-enabled').value === 'true';
    const sslCert = $('#web-ssl-cert').value;
    const sslKey = $('#web-ssl-key').value;

    try {
        await go.StartWebServer(port, registration, sslEnabled, sslCert, sslKey);
        updateWebStatus(true, port, sslEnabled);
        showToast('웹 서버가 시작되었습니다');
    } catch (e) {
        showToast('서버 시작 실패: ' + e, 'error');
    }
}

async function stopWebServer() {
    try {
        await go.StopWebServer();
        showToast('웹 서버가 정지되었습니다');
        updateWebStatus(false);
    } catch (e) { showToast('웹 서버 정지 실패: ' + e, 'error'); }
}

function updateWebStatus(running, port = '8080', ssl = false) {
    const statusEl = $('#web-status-text');
    const urlLink = $('#web-link');
    const addressDisplay = $('#web-address-display');
    const headerDot = $('#header-web-dot');

    if (running) {
        if (statusEl) { statusEl.textContent = '실행 중'; statusEl.className = 'value status-on'; }
        if (urlLink) {
            const protocol = ssl ? 'https' : 'http';
            const url = `${protocol}://localhost:${port}`;
            urlLink.href = url; urlLink.textContent = url;
        }
        if (addressDisplay) addressDisplay.style.display = 'block';
        if (headerDot) headerDot.className = 'status-dot status-on';
    } else {
        if (statusEl) { statusEl.textContent = '정지됨'; statusEl.className = 'value status-off'; }
        if (urlLink) { urlLink.href = '#'; urlLink.textContent = '-'; }
        if (addressDisplay) addressDisplay.style.display = 'none';
        if (headerDot) headerDot.className = 'status-dot status-off';
    }
}

async function loadWebConfig() {
    try {
        const config = await go.GetWebServerConfig();
        if (config) {
            $('#web-port').value = config.port;
            $('#web-registration').value = config.registrationOpen.toString();
            $('#web-ssl-enabled').value = config.sslEnabled.toString();
            $('#web-ssl-cert').value = config.sslCertPath || "";
            $('#web-ssl-key').value = config.sslKeyPath || "";

            // SSL 설정 표시 여부 초기화
            const sslGroup = $('#ssl-settings-group');
            if (config.sslEnabled) {
                sslGroup.style.display = 'block';
            } else {
                sslGroup.style.display = 'none';
            }

            updateWebStatus(config.running, config.port, config.sslEnabled);
        }
    } catch (e) {
        showToast('웹 서버 설정 로드 실패: ' + e, 'error');
    }
}

// 테이블 내 변경 핸들러 (성별 변경 시 아바타 목록 갱신 + 자동 저장)
window.handleTableChange = function (el, field, id) {
    const tr = el.closest('tr');

    if (field === 'gender') {
        const gender = el.value;
        const avatarSelect = tr.querySelector('[data-field="avatar_image"]');
        const avatarPreview = tr.querySelector(`#avatar-preview-${id}`);

        // 아바타 목록 갱신
        const list = state.avatarCache[gender === '남성' ? 'male' : 'female'] || [];
        avatarSelect.innerHTML = '<option value="">(없음)</option>' + list.map(img =>
            `<option value="${img}">${img}</option>`
        ).join('');

        // 미리보기 초기화 (또는 첫번째로 설정)
        if (list.length > 0) {
            avatarSelect.value = list[0];
            avatarPreview.src = `/avarta/${gender === '남성' ? 'male' : 'female'}/${list[0]}`;
        } else {
            avatarSelect.value = "";
            avatarPreview.src = "";
        }

        // 변경 사항 저장 (디바운스)
        debouncedUpdateCharacter(id, tr);
    } else if (field === 'avatar_image') {
        // 아바타 변경 시 미리보기 업데이트 및 저장
        const genderSelect = tr.querySelector('[data-field="gender"]');
        const gender = genderSelect.value;
        const avatarPreview = tr.querySelector(`#avatar-preview-${id}`);

        if (el.value) {
            avatarPreview.src = `/avarta/${gender === '남성' ? 'male' : 'female'}/${el.value}`;
        } else {
            avatarPreview.src = ""; // or placeholder
        }

        // 변경 사항 저장 (디바운스)
        debouncedUpdateCharacter(id, tr);
    }
};

// MBTI 등 다른 input에도 저장 이벤트 연결 필요 (기존엔 없었음)
// 이벤트 위임으로 처리
document.addEventListener('change', function (e) {
    if (e.target.matches('#character-tbody [data-field]')) {
        const field = e.target.dataset.field;
        // gender와 avatar_image는 handleTableChange에서 처리함
        if (field !== 'gender' && field !== 'avatar_image') {
            const tr = e.target.closest('tr');
            if (tr) {
                const id = tr.dataset.charId;
                debouncedUpdateCharacter(id, tr);
            }
        }
    }
});

async function saveWebConfig() {
    const port = $('#web-port').value;
    const registration = $('#web-registration').value === 'true';
    const sslEnabled = $('#web-ssl-enabled').value === 'true';
    const sslCert = $('#web-ssl-cert').value;
    const sslKey = $('#web-ssl-key').value;

    try {
        await go.SaveWebServerConfig(port, registration, sslEnabled, sslCert, sslKey);
        showToast('웹 서버 설정이 저장되었습니다');
    } catch (e) {
        showToast('설정 저장 실패: ' + e, 'error');
    }
}

async function loadBBSConfig() {
    try {
        const config = await go.GetBBSConfig();
        if (config) {
            $('#bbs-title').value = config.title;
            $('#bbs-footer').value = config.footer;
            $('#bbs-theme').value = config.theme;
            $('#bbs-font').value = config.font;
            $('#bbs-posts-per-page').value = config.posts_per_page || 20;
            $('#bbs-timezone').value = config.timezone || 'Asia/Seoul';
        }
    } catch (e) { console.log('BBS 설정 로드 실패', e); }
}

async function saveBBSConfig() {
    const title = $('#bbs-title').value;
    const footer = $('#bbs-footer').value;
    const theme = $('#bbs-theme').value;
    const font = $('#bbs-font').value;
    const postsPerPage = parseInt($('#bbs-posts-per-page').value) || 20;
    const timezone = $('#bbs-timezone').value;

    try {
        await go.SaveBBSConfig(title, footer, theme, font, postsPerPage, timezone);
        showToast('게시판 설정이 저장되었습니다');
    } catch (e) { showToast('설정 저장 실패: ' + e, 'error'); }
}

// ---------------------------------------------------------
// AI 프롬프트 관리
// ---------------------------------------------------------

const PROMPT_MAP = {
    'nickname_gen': 'prompt-nickname',
    'system_role': 'prompt-system',
    'topic_hints': 'prompt-topic-hints',
    'post_instruction': 'prompt-post',
    'comment_instruction': 'prompt-comment',
    'reply_instruction': 'prompt-reply',
    'summary_instruction': 'prompt-summary'
};

async function loadCompPrompts() {
    for (const [key, elementId] of Object.entries(PROMPT_MAP)) {
        try {
            const content = await go.GetPrompt(key);
            $(`#${elementId}`).value = content;
        } catch (e) {
            console.error(`프롬프트 로드 실패 (${key}):`, e);
            $(`#${elementId}`).value = "(로드 실패)";
        }
    }
}

async function saveCompPrompt(key) {
    const elementId = PROMPT_MAP[key];
    const content = $(`#${elementId}`).value;
    try {
        await go.SavePrompt(key, content);
        showToast('프롬프트가 저장되었습니다.');
    } catch (e) {
        showToast('저장 실패: ' + e, 'error');
    }
}

async function resetCompPrompt(key) {
    if (!confirm('정말로 이 프롬프트를 기본값으로 초기화하시겠습니까?')) return;
    try {
        await go.ResetPrompt(key);
        // 다시 로드하여 UI 갱신 (서버에서 기본값 반환)
        const content = await go.GetPrompt(key);
        const elementId = PROMPT_MAP[key];
        $(`#${elementId}`).value = content;
        showToast('기본값으로 초기화되었습니다.');
    } catch (e) {
        showToast('초기화 실패: ' + e, 'error');
    }
}

async function loadMBTIList() {
    try {
        const descs = await go.GetMBTIDescriptions();
        const tbody = $('#mbti-table tbody');
        tbody.innerHTML = '';

        // MBTI 순서 정렬 (알파벳순 등)
        const mbtis = Object.keys(descs).sort();

        mbtis.forEach(mbti => {
            const desc = descs[mbti];
            const row = document.createElement('tr');
            row.innerHTML = `
                <td style="font-weight:bold;text-align:center;">${mbti}</td>
                <td><input type="text" class="form-input input-mbti-desc" value="${desc.replace(/"/g, '&quot;')}" style="width:100%; min-width: 500px;"></td>
                <td style="text-align:center;">
                    <button class="btn-primary btn-save-mbti" data-mbti="${mbti}" style="font-size:12px;padding:4px 8px;">저장</button>
                    <button class="btn-warning btn-reset-mbti" data-mbti="${mbti}" style="font-size:12px;padding:4px 8px;">초기화</button>
                </td>
            `;
            tbody.appendChild(row);
        });
    } catch (e) {
        console.error('MBTI 목록 로드 실패', e);
        showToast('MBTI 목록을 불러오지 못했습니다.', 'error');
    }
}

async function saveMBTI(mbti, content) {
    try {
        await go.SaveMBTIDescription(mbti, content);
        showToast(`${mbti} 설명이 저장되었습니다.`);
    } catch (e) {
        showToast('저장 실패: ' + e, 'error');
    }
}

async function resetMBTI(mbti) {
    if (!confirm(`${mbti} 설명을 기본값으로 초기화하시겠습니까?`)) return;
    try {
        await go.ResetMBTIDescription(mbti);
        // 전체 리로드하는게 속편함
        loadMBTIList();
        showToast('초기화되었습니다.');
    } catch (e) {
        showToast('초기화 실패: ' + e, 'error');
    }
}

// ================================
// 캐릭터 관리자 탭 전환
// ================================
$$('.char-tab-btn').forEach(btn => {
    btn.addEventListener('click', (e) => {
        const tabId = e.target.dataset.charTab;

        // 모든 탭 버튼 비활성화
        $$('.char-tab-btn').forEach(b => b.classList.remove('active'));
        e.target.classList.add('active');

        // 모든 탭 콘텐츠 숨김
        $$('.char-tab-content').forEach(c => c.style.display = 'none');

        // 선택된 탭 콘텐츠 표시
        const content = $(`#char-tab-${tabId}`);
        if (content) {
            content.style.display = 'block';

            // 생성 참조값 탭 선택 시 데이터 로드
            if (tabId === 'refs') {
                loadCharacterRefValues();
            }
        }
    });
});

// ================================
// 캐릭터 생성 참조값 관리
// ================================
// 생성 설정 UI 업데이트 (활성/비활성)
function updateGenSettingsUI() {
    const useAge = $('#chk-use-age').checked;
    const useGender = $('#chk-use-gender').checked;

    // Age Group
    const ageGroup = $('#group-age-settings');
    if (ageGroup) {
        ageGroup.style.opacity = useAge ? '1' : '0.5';
        ageGroup.style.pointerEvents = useAge ? 'auto' : 'none';
    }

    // Gender Group
    const genderGroup = $('#group-gender-settings');
    if (genderGroup) {
        genderGroup.style.opacity = useGender ? '1' : '0.5';
        genderGroup.style.pointerEvents = useGender ? 'auto' : 'none';
    }
}

// 체크박스 이벤트 리스너 추가 (한 번만)
const chkUseAge = $('#chk-use-age');
if (chkUseAge && !chkUseAge.dataset.listenerAttached) {
    chkUseAge.addEventListener('change', updateGenSettingsUI);
    chkUseAge.dataset.listenerAttached = 'true';
}
const chkUseGender = $('#chk-use-gender');
if (chkUseGender && !chkUseGender.dataset.listenerAttached) {
    chkUseGender.addEventListener('change', updateGenSettingsUI);
    chkUseGender.dataset.listenerAttached = 'true';
}

async function loadCharacterRefValues() {
    try {
        const refs = await go.GetCharacterRefValues();
        $('#ref-job-categories').value = refs.job_categories || '';
        $('#ref-hobbies').value = refs.hobbies || '';
        $('#ref-regions').value = refs.regions || '';

        // 생성 기본 설정 로드
        const settings = await go.GetGenSettings();
        if (settings) {
            if ($('#ref-min-age')) $('#ref-min-age').value = settings.min_age;
            if ($('#ref-max-age')) $('#ref-max-age').value = settings.max_age;
            if ($('#ref-male-ratio')) {
                $('#ref-male-ratio').value = settings.male_ratio;
                $('#ref-female-ratio-display').textContent = 100 - settings.male_ratio;
            }
            if ($('#chk-use-age')) $('#chk-use-age').checked = settings.use_age;
            if ($('#chk-use-gender')) $('#chk-use-gender').checked = settings.use_gender;

            updateGenSettingsUI();
        }
    } catch (e) {
        console.error('참조값 로드 실패', e);
    }
}

// 생성 기본 설정 저장 버튼
const btnSaveGenSettings = $('#btn-save-gen-settings');
if (btnSaveGenSettings) {
    btnSaveGenSettings.addEventListener('click', async () => {
        const minAge = parseInt($('#ref-min-age').value) || 15;
        const maxAge = parseInt($('#ref-max-age').value) || 64;
        const maleRatio = parseInt($('#ref-male-ratio').value) || 50;
        const useAge = $('#chk-use-age').checked;
        const useGender = $('#chk-use-gender').checked;

        if (useAge && minAge > maxAge) {
            showToast('최소 나이가 최대 나이보다 클 수 없습니다.', 'error');
            return;
        }
        if (useGender && (maleRatio < 0 || maleRatio > 100)) {
            showToast('성비는 0~100 사이여야 합니다.', 'error');
            return;
        }

        try {
            await go.SaveGenSettings(minAge, maxAge, maleRatio, useAge, useGender);
            showToast('기본 설정이 저장되었습니다.');
        } catch (e) {
            showToast('설정 저장 실패: ' + e, 'error');
        }
    });

    // 성비 자동 계산
    $('#ref-male-ratio').addEventListener('input', (e) => {
        let val = parseInt(e.target.value);
        if (isNaN(val)) val = 0;
        if (val < 0) val = 0;
        if (val > 100) val = 100;

        $('#ref-female-ratio-display').textContent = 100 - val;
    });
}

// 참조값 저장 버튼
$$('.btn-save-ref').forEach(btn => {
    btn.addEventListener('click', async () => {
        const key = btn.dataset.key;
        const textareaId = `#ref-${key.replace('_', '-')}`;
        const value = $(textareaId).value.trim();

        try {
            await go.SaveCharacterRefValue(key, value);
            showToast('저장되었습니다.');
        } catch (e) {
            showToast('저장 실패: ' + e, 'error');
        }
    });
});

// 참조값 초기화 버튼
$$('.btn-reset-ref').forEach(btn => {
    btn.addEventListener('click', async () => {
        const key = btn.dataset.key;
        const textareaId = `#ref-${key.replace('_', '-')}`;

        if (!confirm('정말로 초기화하시겠습니까?')) {
            return;
        }

        try {
            const defaultValue = await go.ResetCharacterRefValue(key);
            $(textareaId).value = defaultValue;
            showToast('초기화되었습니다.');
        } catch (e) {
            showToast('초기화 실패: ' + e, 'error');
        }
    });
});

// ================================
// 캐릭터 일괄 선택 및 작업
// ================================

// 헤더 체크박스로 현재 보이는 행 전체 선택/해제
const selectAllCheckbox = $('#char-select-all-checkbox');
if (selectAllCheckbox) {
    selectAllCheckbox.addEventListener('change', () => {
        const checkboxes = $$('.char-select-checkbox');
        checkboxes.forEach(cb => cb.checked = selectAllCheckbox.checked);
    });
}

// 전체 선택 버튼
const btnSelectAll = $('#btn-select-all');
if (btnSelectAll) {
    btnSelectAll.addEventListener('click', () => {
        const checkboxes = $$('.char-select-checkbox');
        const allChecked = [...checkboxes].every(cb => cb.checked);
        checkboxes.forEach(cb => cb.checked = !allChecked);
        if (selectAllCheckbox) selectAllCheckbox.checked = !allChecked;
    });
}

// 선택된 캐릭터 ID 목록 가져오기
function getSelectedCharacterIds() {
    const checked = $$('.char-select-checkbox:checked');
    return [...checked].map(cb => parseInt(cb.dataset.charId));
}

// 일괄 활성화
const btnBatchActivate = $('#btn-batch-activate');
if (btnBatchActivate) {
    btnBatchActivate.addEventListener('click', async () => {
        const ids = getSelectedCharacterIds();
        if (ids.length === 0) {
            showToast('선택된 캐릭터가 없습니다.', 'error');
            return;
        }
        try {
            await go.BatchSetCharacterActive(ids, true);
            showToast(`${ids.length}명 활성화 완료`);
            loadCharacters();
        } catch (e) {
            showToast('활성화 실패: ' + e, 'error');
        }
    });
}

// 일괄 비활성화
const btnBatchDeactivate = $('#btn-batch-deactivate');
if (btnBatchDeactivate) {
    btnBatchDeactivate.addEventListener('click', async () => {
        const ids = getSelectedCharacterIds();
        if (ids.length === 0) {
            showToast('선택된 캐릭터가 없습니다.', 'error');
            return;
        }
        try {
            await go.BatchSetCharacterActive(ids, false);
            showToast(`${ids.length}명 비활성화 완료`);
            loadCharacters();
        } catch (e) {
            showToast('비활성화 실패: ' + e, 'error');
        }
    });
}

// 일괄 삭제
const btnBatchDelete = $('#btn-batch-delete');
if (btnBatchDelete) {
    btnBatchDelete.addEventListener('click', async () => {
        const ids = getSelectedCharacterIds();
        console.log('[DEBUG] Selected character IDs:', ids);
        if (ids.length === 0) {
            showToast('선택된 캐릭터가 없습니다.', 'error');
            return;
        }
        if (!confirm(`${ids.length}명의 캐릭터를 삭제하시겠습니까?`)) {
            return;
        }
        try {
            await go.BatchDeleteCharacters(ids);
            showToast(`${ids.length}명 삭제 완료`);
            // DB 갱신 반영을 위해 약간의 지연 후 로드
            setTimeout(async () => {
                await loadCharacters();
                console.log('[DEBUG] Refreshed characters after delete');
            }, 100);
        } catch (e) {
            showToast('삭제 실패: ' + e, 'error');
        }
    });
}

// 전역 함수 노출 (HTML onclick 이벤트용)
// 전역 함수 노출 (HTML onclick 이벤트용)
window.handleCharacterSort = handleCharacterSort;

// ================================
// 참조값 파일 불러오기 / 내보내기
// ================================
const btnImportRef = $('#btn-import-ref');
if (btnImportRef) {
    btnImportRef.addEventListener('click', async () => {
        try {
            const result = await go.ImportReferenceValues();
            if (result) {
                // 경로 구분자 처리 (Windows/Unix)
                const filename = result.split(/[\\/]/).pop();
                showToast(`참조값을 성공적으로 불러왔습니다: ${filename}`);
                loadCharacterRefValues(); // UI 갱신
            }
        } catch (e) {
            showToast('불러오기 실패: ' + e, 'error');
        }
    });
}

const btnExportRef = $('#btn-export-ref');
if (btnExportRef) {
    btnExportRef.addEventListener('click', async () => {
        try {
            const result = await go.ExportReferenceValues();
            if (result) {
                const filename = result.split(/[\\/]/).pop();
                showToast(`참조값을 성공적으로 저장했습니다: ${filename}`);
            }
        } catch (e) {
            showToast('내보내기 실패: ' + e, 'error');
        }
    });
}

// ================================
// 프롬프트 설정 불러오기 / 내보내기
// ================================
const btnImportPrompt = $('#btn-import-prompt');
if (btnImportPrompt) {
    btnImportPrompt.addEventListener('click', async () => {
        try {
            const result = await go.ImportPromptSettings();
            if (result) {
                const filename = result.split(/[\\/]/).pop();
                showToast(`프롬프트를 성공적으로 불러왔습니다: ${filename}`);
                loadCompPrompts();
                loadMBTIList();
            }
        } catch (e) {
            showToast('불러오기 실패: ' + e, 'error');
        }
    });
}

const btnExportPrompt = $('#btn-export-prompt');
if (btnExportPrompt) {
    btnExportPrompt.addEventListener('click', async () => {
        try {
            const result = await go.ExportPromptSettings();
            if (result) {
                const filename = result.split(/[\\/]/).pop();
                showToast(`프롬프트를 성공적으로 저장했습니다: ${filename}`);
            }
        } catch (e) {
            showToast('내보내기 실패: ' + e, 'error');
        }
    });
}
