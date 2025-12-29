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

// 초기화
document.addEventListener('DOMContentLoaded', async () => {
    initEventListeners();
    await checkLoginStatus();
    await checkDBStatus();

    // 초기 데이터 로드
    loadWebConfig();
    loadLLMConfig();
    loadBBSConfig();
});

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

            if (!state.currentUser || !state.currentUser.is_admin) {
                showToast('관리자 권한이 필요합니다.', 'error');
                checkbox.checked = !newStatus;
                return;
            }

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

function escapeHtml(text) {
    if (!text) return '';
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
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

    const isAdmin = state.currentUser && state.currentUser.is_admin;

    users.forEach(user => {
        const tr = document.createElement('tr');
        const disabledAttr = isAdmin ? '' : 'disabled';

        tr.innerHTML = `
            <td>${user.id}</td>
            <td>${escapeHtml(user.username)}</td>
            <td>${escapeHtml(user.nickname)}</td>
            <td>
                <input type="checkbox" class="admin-check" data-id="${user.id}" 
                    ${user.is_admin ? 'checked' : ''} ${disabledAttr} />
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
        }
    } catch (e) { console.log('DB 상태 확인 실패:', e); }
}

async function resetDatabase() {
    const confirmation = $('#confirm-input').value;
    try {
        await go.ResetDatabase(confirmation);
        showToast('데이터베이스가 초기화되었습니다');
        hideModal('confirm-modal');
        $('#confirm-input').value = '';
        loadUsers();
        const chars = await go.GetAllCharacters();
        renderCharacters(chars);
    } catch (e) { showToast('초기화 실패: ' + e, 'error'); }
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

    try {
        await go.SaveLLMConfig(host, port, model1, model2, model3, postsPerHour, commentsPerHour, maxTokens, temperature);
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
        renderCharacters(chars);
    } catch (e) { showToast('캐릭터 생성 실패: ' + e, 'error'); }
}

async function loadCharacters() {
    try {
        const chars = await go.GetAllCharacters();
        const stats = await go.GetAllCharacterStats();
        state.cachedCharacters = chars;
        state.cachedStats = stats;
        sortAndRenderCharacters();
    } catch (e) { console.log('캐릭터 로딩 실패:', e); }
}

function sortAndRenderCharacters() {
    const { field, asc } = state.characterSort;
    const chars = [...state.cachedCharacters];
    const stats = state.cachedStats;

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

    renderCharacters(chars, stats);
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

function renderCharacters(chars, stats = {}) {
    const tbody = $('#character-tbody');
    tbody.innerHTML = '';
    if (!chars || chars.length === 0) {
        tbody.innerHTML = '<tr><td colspan="16" style="text-align:center;padding:20px;">캐릭터가 없습니다.</td></tr>';
        return;
    }
    chars.forEach(c => {
        const charStats = stats[c.id] || { post_count: 0, comment_count: 0 };
        const tr = document.createElement('tr');
        tr.innerHTML = `
            <td>${c.id}</td>
            <td><input type="text" value="${escapeHtml(c.nickname)}" data-field="nickname" style="width:80px;"/></td>
            <td>
                <select data-field="gender">
                    <option value="남성" ${c.gender === '남성' ? 'selected' : ''}>남성</option>
                    <option value="여성" ${c.gender === '여성' ? 'selected' : ''}>여성</option>
                </select>
            </td>
            <td><input type="number" value="${c.age}" min="10" max="80" data-field="age" style="width:40px;"/></td>
            <td><input type="text" value="${c.birthdate || ''}" data-field="birthdate" placeholder="YYYY-MM-DD" style="width:90px;"/></td>
            <td><input type="text" value="${escapeHtml(c.region) || ''}" data-field="region" style="width:60px;"/></td>
            <td><input type="text" value="${escapeHtml(c.job_category)}" data-field="job_category" style="width:100px;"/></td>
            <td><input type="text" value="${c.mbti}" maxlength="4" data-field="mbti" style="width:50px;"/></td>
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
        id: id,
        nickname: tr.querySelector('[data-field="nickname"]').value,
        gender: tr.querySelector('[data-field="gender"]').value,
        age: parseInt(tr.querySelector('[data-field="age"]').value),
        birthdate: tr.querySelector('[data-field="birthdate"]').value,
        region: tr.querySelector('[data-field="region"]').value,
        job_category: tr.querySelector('[data-field="job_category"]').value,
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

async function deleteCharacter(id) {
    if (!confirm('정말 삭제하시겠습니까?')) return;
    try {
        await go.DeleteCharacter(id);
        showToast('캐릭터가 삭제되었습니다');
        loadCharacters();
    } catch (e) { showToast('캐릭터 삭제 실패: ' + e, 'error'); }
}

async function startWebServer() {
    const port = $('#web-port').value || '8080';
    const regOpen = $('#web-registration').value === 'true';
    try {
        await go.StartWebServer(port, regOpen);
        showToast(`웹 서버가 포트 ${port}에서 시작되었습니다`);
        updateWebStatus(true, port);
    } catch (e) { showToast('웹 서버 시작 실패: ' + e, 'error'); }
}

async function stopWebServer() {
    try {
        await go.StopWebServer();
        showToast('웹 서버가 정지되었습니다');
        updateWebStatus(false);
    } catch (e) { showToast('웹 서버 정지 실패: ' + e, 'error'); }
}

function updateWebStatus(running, port = '8080') {
    const statusEl = $('#web-status-text');
    const urlLink = $('#web-link');
    const portDisplay = $('#web-port-display');
    const headerDot = $('#header-web-dot');

    if (running) {
        if (statusEl) { statusEl.textContent = '실행 중'; statusEl.className = 'value status-on'; }
        if (urlLink) {
            const url = `http://localhost:${port}`;
            urlLink.href = url; urlLink.textContent = url;
        }
        if (portDisplay) portDisplay.textContent = port;
        if (headerDot) headerDot.className = 'status-dot status-on';
    } else {
        if (statusEl) { statusEl.textContent = '정지됨'; statusEl.className = 'value status-off'; }
        if (urlLink) { urlLink.href = '#'; urlLink.textContent = '-'; }
        if (portDisplay) portDisplay.textContent = '-';
        if (headerDot) headerDot.className = 'status-dot status-off';
    }
}

async function loadWebConfig() {
    try {
        const config = await go.GetWebServerConfig();
        if (config) {
            $('#web-port').value = config.port;
            $('#web-registration').value = config.registrationOpen.toString();
            updateWebStatus(config.running, config.port);
        }
    } catch (e) { console.log('웹 설정 로드 실패', e); }
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

// 전역 함수 노출 (HTML onclick 이벤트용)
window.handleCharacterSort = handleCharacterSort;
