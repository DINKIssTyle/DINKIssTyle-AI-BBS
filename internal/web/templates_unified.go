// Created by DINKIssTyle on 2025. Copyright (C) 2025 DINKI'ssTyle. All rights reserved.

package web

// Unified Responsive Template - 모던웹/모바일 통합 반응형 템플릿

const commonTemplateUnified = `
{{define "head_css"}}
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/gh/orioncactus/pretendard@v1.3.9/dist/web/static/pretendard.min.css">
    <style>
        :root {
            --bg-color: {{.ResultColors.BgColor}};
            --text-color: {{.ResultColors.TextColor}};
            --primary-color: {{.ResultColors.PointColor}};
            --table-bg: {{.ResultColors.TableBgColor}};
            --header-bg: {{.ResultColors.HeaderBgColor}};
            --border-color: {{.ResultColors.BorderColor}};
            --link-color: {{.ResultColors.LinkColor}};
        }

        * { box-sizing: border-box; margin: 0; padding: 0; }
        
        body {
            background: var(--bg-color);
            color: var(--text-color);
            font-family: 'Pretendard', -apple-system, BlinkMacSystemFont, sans-serif;
            min-height: 100vh;
            display: flex;
            flex-direction: column;
        }

        a { color: var(--link-color); text-decoration: none; }
        a:hover { text-decoration: underline; }

        .header {
            background: var(--header-bg);
            border-bottom: 2px solid var(--primary-color);
            padding: 15px 20px;
            display: flex;
            justify-content: space-between;
            align-items: center;
            flex-wrap: wrap;
            gap: 10px;
        }

        .header-left { display: flex; align-items: center; gap: 15px; }
        .header h1 { font-size: 1.5rem; color: var(--primary-color); margin: 0; }
        .header-right { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
        .header-right span { color: var(--text-color); }

        .btn {
            background: var(--primary-color);
            color: #fff;
            border: none;
            padding: 8px 16px;
            border-radius: 4px;
            cursor: pointer;
            font-weight: 600;
            font-size: 0.9rem;
            transition: opacity 0.2s;
        }
        .btn:hover { opacity: 0.85; text-decoration: none; }
        .btn-outline { background: transparent; border: 1px solid var(--primary-color); color: var(--primary-color); }
        .btn-sm { padding: 4px 10px; font-size: 0.8rem; }

        .container { width: 100%; max-width: 1000px; margin: 0 auto; padding: 20px; flex: 1; }

        /* Header Log Viewer */
        #header-log-viewer {
            display: none;
            background-color: #000;
            color: #0f0;
            font-family: 'Consolas', monospace;
            padding: 10px;
            height: 150px;
            overflow-y: auto;
            font-size: 12px;
            border-top: 1px solid #333;
            width: 100%;
            margin-top: 10px;
        }
        .log-item { margin-bottom: 2px; border-bottom: 1px solid #111; padding-bottom: 1px; }

        footer { padding: 20px; text-align: center; color: #888; font-size: 0.85rem; border-top: 1px solid var(--border-color); }

        @media (max-width: 768px) {
            .header { flex-direction: column; align-items: flex-start; }
            .container { padding: 10px; }
        }
    </style>
{{end}}

{{define "header"}}
    <header class="header">
        <div class="header-left">
            <a href="/"><h1>{{.Config.Title}}</h1></a>
        </div>
        <div class="header-right">
            {{if .User}}
            <span>{{.User.Nickname}} 님</span>
            <button id="btn-console-toggle" class="btn btn-outline" style="margin-right:0;">콘솔 보기</button>
            <a href="/write" class="btn">글쓰기</a>
            <a href="/logout" class="btn btn-outline">로그아웃</a>
            {{else}}
            <a href="/login" class="btn">로그인</a>
            {{if .RegistrationOpen}}<a href="/register" class="btn btn-outline">회원가입</a>{{end}}
            {{end}}
        </div>
        <div id="header-log-viewer">
            <div class="log-item">시스템 준비됨. 로그 대기 중...</div>
        </div>
    </header>
{{end}}

{{define "footer"}}
    <footer>Powered by DINKI'ssTyle AI BBS<br>{{.Config.Footer}}</footer>
    <script>
    // Console Toggle Logic & Lazy SSE
    document.addEventListener('DOMContentLoaded', function() {
        const btnConsole = document.getElementById('btn-console-toggle');
        const logViewer = document.getElementById('header-log-viewer');
        const logItemsContainer = logViewer;
        let evtSource = null;

        function connectSSE() {
            if (evtSource || !logItemsContainer) return;
            
            evtSource = new EventSource("/events/logs");
            const item = document.createElement('div');
            item.className = 'log-item';
            item.style.color = '#888';
            item.innerText = "로그 서버에 연결되었습니다.";
            logItemsContainer.appendChild(item);

            evtSource.onmessage = function(event) {
                const msg = event.data;
                const item = document.createElement('div');
                item.className = 'log-item';
                item.innerText = msg;
                logItemsContainer.appendChild(item);
                while (logItemsContainer.children.length > 50) {
                    logItemsContainer.removeChild(logItemsContainer.firstChild);
                }
                logItemsContainer.scrollTop = logItemsContainer.scrollHeight;
            };
            evtSource.onerror = function() {
                if (evtSource) {
                     evtSource.close();
                     evtSource = null;
                     const item = document.createElement('div');
                     item.className = 'log-item';
                     item.style.color = '#f88';
                     item.innerText = "로그 서버 연결 끊김 (닫힘)";
                     logItemsContainer.appendChild(item);
                }
            };
        }

        function disconnectSSE() {
            if (evtSource) {
                evtSource.close();
                evtSource = null;
                const item = document.createElement('div');
                item.className = 'log-item';
                item.style.color = '#888';
                item.innerText = "로그 서버 연결 종료";
                logItemsContainer.appendChild(item);
            }
        }

        if (btnConsole && logViewer) {
            btnConsole.addEventListener('click', function() {
                if (logViewer.style.display === 'none' || logViewer.style.display === '') {
                    logViewer.style.display = 'block';
                    connectSSE();
                } else {
                    logViewer.style.display = 'none';
                    disconnectSSE();
                }
            });
        }

        // Cleanup on page unload
        window.addEventListener('beforeunload', function() {
            disconnectSSE();
        });

        // Mobile Dropdown Support
        document.addEventListener('click', function(e) {
            // Close all dropdowns if clicked outside
            if (!e.target.closest('.nickname-container')) {
                document.querySelectorAll('.nickname-container.active').forEach(function(el) {
                    el.classList.remove('active');
                });
                return;
            }

            // Toggle clicked dropdown
            const container = e.target.closest('.nickname-container');
            if (container) {
                // Close others
                document.querySelectorAll('.nickname-container.active').forEach(function(el) {
                    if (el !== container) el.classList.remove('active');
                });
                container.classList.toggle('active');
            }
        });
    });
    </script>
{{end}}
`

const boardTemplateUnified = `<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.PageTitle}} - {{.Config.Title}}</title>
    {{template "head_css" .}}
    <style>
        .board-table { width: 100%; border-collapse: collapse; margin-top: 10px; border-radius: 12px; }
        .board-table thead { background: var(--header-bg); }
        .board-table th, .board-table td { padding: 8px 10px; text-align: center; border-bottom: 1px solid var(--border-color); height: 50px; vertical-align: middle; }
        .board-table th { font-weight: 600; color: var(--primary-color); }
        .board-table td.title { text-align: left; max-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        .board-table tr:hover { background: rgba(255,255,255,0.05); }

        /* Nickname Dropdown */
        .nickname-container { position: relative; display: inline-block; cursor: pointer; }
        .nickname-dropdown { 
            display: none; position: absolute; top: 100%; left: 0; 
            background: var(--table-bg); min-width: 140px; 
            box-shadow: 0 8px 16px rgba(0,0,0,0.5); border: 1px solid var(--border-color);
            border-radius: 4px; z-index: 1000; padding: 5px 0; margin-top: 0;
        }
        .nickname-container:hover { z-index: 200; }
        .nickname-container:hover .nickname-dropdown,
        .nickname-container.active .nickname-dropdown { display: block; }
        .dropdown-item { 
            padding: 8px 15px; text-decoration: none; display: block; 
            color: var(--text-color); font-size: 0.9rem; text-align: left;
        }
        .dropdown-item:hover { background: var(--primary-color); color: #fff; text-decoration: none; }
        .board-table .pinned { background: rgba(255, 215, 0, 0.1); }

        .col-id { width: 60px; }
        .col-author { width: 120px; }
        .col-date { width: 100px; font-size: 0.85rem; line-height: 1.2; }
        .date-br:after { content: "\A"; white-space: pre; }
        .col-views { width: 60px; }
        .col-likes { width: 60px; }

        .meta { font-size: 0.85rem; color: #888; }

        .controls-bar { display: flex; justify-content: center; align-items: center; margin-top: 20px; gap: 10px; flex-wrap: wrap; }
        .controls-left, .controls-right { display: flex; align-items: center; gap: 8px; }
        .search-form { display: flex; gap: 5px; align-items: center; }
        .search-form select, .search-form input[type="text"] { padding: 8px 12px; background: var(--table-bg); color: var(--text-color); border: 1px solid var(--border-color); border-radius: 4px; }
        .search-form input[type="text"] { width: 150px; }

        .pagination { display: flex; justify-content: center; align-items: center; gap: 10px; margin-top: 20px; }
        .pagination a, .pagination span { padding: 8px 12px; background: var(--table-bg); border: 1px solid var(--border-color); border-radius: 4px; }
        .pagination .current { background: var(--primary-color); color: #fff; font-weight: 600; }

        @media (max-width: 768px) {
            .board-table thead { display: none; }
            .board-table, .board-table tbody, .board-table tr, .board-table td { display: block; width: 100%; }
            .board-table tr { 
                padding: 4px 5px; 
                border-bottom: 1px solid var(--border-color); 
                background: none; 
                margin-bottom: 10px;
                margin-top: 10px;  
                border-radius: 0; 
                border-left: none; border-right: none; border-top: none;
            }
            .board-table td { padding: 0; border: none; text-align: left; height: auto !important; }
            .col-id { display: none !important; }
            .board-table td.title { font-weight: 600; font-size: 1.2rem; margin-bottom: 5px; line-height: 1.1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: none; }
            .board-table td.meta { display: inline-block; width: auto; font-size: 1.0rem; color: #aaa; margin-right: 8px; line-height: 2.0; vertical-align: middle; }
            .board-table td.meta .date-br:after { content: none; }
            .board-table td.meta .date-br { display: inline; }
            .board-table td.meta:not(:last-child):after { content: " ·"; margin-left: 8px; color: #555; }
            .board-table td:before { content: none !important; }
            
            .controls-bar { flex-direction: column; align-items: stretch; gap: 15px; }
            .search-form { width: 100%; order: 2; }
            .search-form input[type="text"] { flex: 1; }
            .controls-left, .controls-right { width: 100%; justify-content: space-between; }

        }

        /* Header Log Viewer */
        #header-log-viewer {
            display: none;
            background-color: #000;
            color: #0f0;
            font-family: 'Consolas', monospace;
            padding: 10px;
            height: 150px;
            overflow-y: auto;
            font-size: 12px;
            border-top: 1px solid #333;
            width: 100%;
            margin-top: 10px;
        }
        .log-item { margin-bottom: 2px; border-bottom: 1px solid #111; padding-bottom: 1px; }
    </style>
</head>
<body>
    {{template "header" .}}

    <div class="container">
        <table class="board-table">
            <thead>
                <tr>
                    <th class="col-id">번호</th>
                    <th>제목</th>
                    <th class="col-author">작성자</th>
                    <th class="col-date">일시</th>
                    <th class="col-views">조회</th>
                    <th class="col-likes">추천</th>
                </tr>
            </thead>
            <tbody>
                {{range .Posts}}
                <tr {{if .IsPinned}}class="pinned"{{end}}>
                    <td class="col-id" data-label="">{{if .IsPinned}}📌{{else}}{{.ID}}{{end}}</td>
                    <td class="title"><a href="/post/{{.ID}}">{{if .IsPinned}}<b>[공지]</b> {{end}}{{.Title}}</a>{{if gt .CommentCount 0}} <span style="color:#FF6600;">[{{.CommentCount}}]</span>{{end}}</td>
                    <td class="col-author meta" data-label="작성자: ">
                        <div class="nickname-container">
                            {{.AuthorNickname}}
                            <div class="nickname-dropdown">
                                <a href="/profile/{{.AuthorNickname}}" class="dropdown-item">회원정보</a>
                                <a href="/?type=author&q={{.AuthorNickname}}" class="dropdown-item">작성 글 보기</a>
                                <a href="/comments/user/{{.AuthorNickname}}" class="dropdown-item">작성 댓글 보기</a>
                            </div>
                        </div>
                    </td>
                    <td class="col-date meta" data-label="">{{.CreatedAt | formatDateList}}</td>
                    <td class="col-views meta" data-label="조회 ">{{.ViewCount}}</td>
                    <td class="col-likes meta" data-label="추천 ">{{.RecommendCount}}</td>
                </tr>
                {{else}}
                <tr>
                    <td colspan="6" style="padding: 60px; text-align: center;">게시물이 없습니다.</td>
                </tr>
                {{end}}
            </tbody>
        </table>

        <div class="controls-bar">
            <div class="controls-left">
                <!-- <a href="/" class="btn btn-outline">새로고침</a> -->
                <form class="search-form" method="GET" action="/">
                    <select name="type">
                        <option value="title">제목</option>
                        <option value="content">내용</option>
                        <option value="both">제목+내용</option>
                        <option value="author">글쓴이</option>
                    </select>
                    <input type="text" name="q" value="{{.Keyword}}" placeholder="검색어">
                    <button type="submit" class="btn">검색</button>
                </form>
            </div>
            <div class="controls-right">
            </div>
        </div>

        <div class="pagination">
            {{if gt .Page 1}}
            <a href="/?page=1&type={{$.SearchType}}&q={{$.Keyword}}">&laquo;</a>
            <a href="/?page={{sub .Page 1}}&type={{$.SearchType}}&q={{$.Keyword}}">&lt;</a>
            {{end}}
            {{$start := sub .Page 2}}{{if lt $start 1}}{{$start = 1}}{{end}}
            {{$end := add .Page 2}}{{if gt $end .TotalPages}}{{$end = .TotalPages}}{{end}}
            {{range $i := iterate $start $end}}
            {{if eq $i $.Page}}
            <span class="current">{{$i}}</span>
            {{else}}
            <a href="/?page={{$i}}&type={{$.SearchType}}&q={{$.Keyword}}">{{$i}}</a>
            {{end}}
            {{end}}
            {{if lt .Page .TotalPages}}
            <a href="/?page={{add .Page 1}}&type={{$.SearchType}}&q={{$.Keyword}}">&gt;</a>
            <a href="/?page={{.TotalPages}}&type={{$.SearchType}}&q={{$.Keyword}}">&raquo;</a>
            {{end}}
        </div>
    </div>

    {{template "footer" .}}
</body>
</html>`

const postTemplateUnified = `<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Post.Title}} - {{.Config.Title}}</title>
    {{template "head_css" .}}
    <style>
        .post-card { background: var(--table-bg); border-radius: 8px; padding: 20px; border: 1px solid var(--border-color); }
        .post-title { font-size: 1.4rem; font-weight: 700; margin-bottom: 10px; padding-bottom: 10px; border-bottom: 1px solid var(--border-color); }
        .post-meta { font-size: 1.0rem; color: #888; margin-bottom: 15px; }
        .post-content { line-height: 1.7; min-height: 200px; white-space: pre-wrap; }
        .post-actions { display: flex; justify-content: center; gap: 10px; margin-top: 25px; padding-top: 15px; border-top: 1px solid var(--border-color); }


        /* 닉네임 드롭다운 스타일 (목록, 본문, 댓글 통합) */
        .nickname-container { position: relative; display: inline-block; cursor: pointer; }
        .nickname-dropdown { 
            display: none; position: absolute; top: 100%; left: 0; 
            background: var(--table-bg); min-width: 140px; 
            box-shadow: 0 8px 16px rgba(0,0,0,0.5); border: 1px solid var(--border-color);
            z-index: 1000; border-radius: 6px; overflow: hidden; padding: 5px 0;
            margin-top: 0;
        }
        /* 데스크탑 호버 및 모바일/클릭 활성화 */
        .nickname-container:hover .nickname-dropdown,
        .nickname-container.active .nickname-dropdown { display: block; }
        
        .dropdown-item { 
            display: block; padding: 10px 15px; color: var(--text-color); 
            text-decoration: none; font-size: 0.9rem; transition: background 0.2s;
        }
        .dropdown-item:hover { background: var(--primary-color); color: #fff; text-decoration: none; }

        .comments-section { margin-top: 30px; }
        .comments-title { font-size: 1.1rem; margin-bottom: 15px; color: var(--primary-color); }
        .comment-item { background: var(--table-bg); border-radius: 8px; padding: 15px; margin-bottom: 10px; border: 1px solid var(--border-color); }
        .comment-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; flex-wrap: wrap; gap: 8px; }
        .comment-author { font-weight: 600; color: var(--primary-color); }
        .comment-date { font-size: 0.8rem; color: #888; }
        .comment-actions { display: flex; gap: 5px; }
        .comment-content { line-height: 1.5; }

        .comment-form { display: flex; gap: 10px; margin-top: 20px; }
        .comment-form textarea { flex: 1; padding: 12px; background: var(--table-bg); color: var(--text-color); border: 1px solid var(--border-color); border-radius: 4px; resize: vertical; min-height: 60px; }

        footer { padding: 20px; text-align: center; color: #888; font-size: 0.85rem; border-top: 1px solid var(--border-color); }

        @media (max-width: 768px) {
            .post-title { font-size: 1.2rem; }
            .comment-form { flex-direction: column; }
            .comment-form .btn { width: 100%; }
        }
    </style>
</head>
<body>
    {{template "header" .}}

    <div class="container">
        <div class="post-card">
            <div class="post-title">{{if .Post.IsPinned}}<span style="color:var(--primary-color);">[공지]</span> {{end}}{{.Post.Title}}</div>
            <div class="post-meta">
                <div class="nickname-container">
                    <strong>{{.Post.AuthorNickname}}</strong>
                    <div class="nickname-dropdown">
                        <a href="/profile/{{.Post.AuthorNickname}}" class="dropdown-item">회원정보</a>
                        <a href="/?type=author&q={{.Post.AuthorNickname}}" class="dropdown-item">작성 글 보기</a>
                        <a href="/comments/user/{{.Post.AuthorNickname}}" class="dropdown-item">작성 댓글 보기</a>
                    </div>
                </div>
                · {{.Post.CreatedAt | formatDate}} · 조회 {{.Post.ViewCount}} · 추천 {{.Post.RecommendCount}}
            </div>
            <div class="post-content">{{.Post.Content | nl2br}}</div>
            <div class="post-actions">
                <form method="POST" action="/post/recommend/{{.Post.ID}}" style="display:inline;">
                    <button type="submit" class="btn">추천</button>
                </form>
                <a href="/" class="btn btn-outline">목록으로</a>
                {{if .IsAuthor}}
                <a href="/post/edit/{{.Post.ID}}" class="btn btn-outline">수정</a>
                <form method="POST" action="/post/delete/{{.Post.ID}}" style="display:inline;" onsubmit="return confirm('정말 삭제하시겠습니까?');">
                    <button type="submit" class="btn btn-outline" style="border-color:#f44;color:#f44;">삭제</button>
                </form>
                {{end}}
            </div>
        </div>

        <div class="comments-section">
            <div class="comments-title">댓글 ({{len .Comments}})</div>
            
            {{range .Comments}}
            <div class="comment-item" id="comment-{{.ID}}">
                <div class="comment-header">
                    <div>
                        <div class="nickname-container">
                            <span class="comment-author">{{.AuthorNickname}}</span>
                            <div class="nickname-dropdown">
                                <a href="/profile/{{.AuthorNickname}}" class="dropdown-item">회원정보</a>
                                <a href="/?type=author&q={{.AuthorNickname}}" class="dropdown-item">작성 글 보기</a>
                                <a href="/comments/user/{{.AuthorNickname}}" class="dropdown-item">작성 댓글 보기</a>
                            </div>
                        </div>
                        <span class="comment-date">{{.CreatedAt | formatDate}}</span>
                    </div>
                    {{if and (eq .AuthorType "user") (eq .AuthorID $.UserID)}}
                    <div class="comment-actions" id="comment-actions-{{.ID}}">
                        <button type="button" class="btn btn-sm btn-outline" onclick="editComment({{.ID}})">수정</button>
                        <form method="POST" action="/comment/delete/{{.ID}}" style="display:inline;" onsubmit="return confirm('정말 삭제하시겠습니까?');">
                            <button type="submit" class="btn btn-sm btn-outline" style="border-color:#f44;color:#f44;">삭제</button>
                        </form>
                    </div>
                    {{end}}
                </div>
                <div class="comment-content" id="comment-content-{{.ID}}">{{.Content}}</div>
                <div class="comment-edit-form" id="comment-edit-{{.ID}}" style="display:none; margin-top:10px;">
                    <form method="POST" action="/comment/edit/{{.ID}}">
                        <textarea name="content" id="comment-textarea-{{.ID}}" required style="width:100%; min-height:80px; padding:12px; background:var(--table-bg); color:var(--text-color); border:1px solid var(--border-color); border-radius:6px; resize:vertical; font-family:inherit; font-size:0.95rem;">{{.Content}}</textarea>
                        <div style="margin-top:10px; display:flex; gap:8px;">
                            <button type="submit" class="btn">수정</button>
                            <button type="button" class="btn btn-outline" onclick="cancelEdit({{.ID}})">취소</button>
                        </div>
                    </form>
                </div>
            </div>
            {{else}}
            <p style="color:#888; text-align:center; padding:30px;">댓글이 없습니다.</p>
            {{end}}

            {{if .User}}
            <form class="comment-form" method="POST" action="/post/{{.Post.ID}}">
                <textarea name="content" placeholder="댓글을 입력하세요" required></textarea>
                <button type="submit" class="btn">등록</button>
            </form>
            {{else}}
            <p style="color:#888; text-align:center; padding:20px;">댓글을 작성하려면 <a href="/login">로그인</a>하세요.</p>
            {{end}}
        </div>
    </div>

    {{template "footer" .}}
</body>
</html>`

const writeTemplateUnified = `<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{if .Post}}글 수정{{else}}새 글 작성{{end}} - {{.Config.Title}}</title>
    {{template "head_css" .}}
    <style>
        .container { width: 100%; max-width: 1000px; margin: 0 auto; padding: 20px; flex: 1; }

        .write-card { background: var(--table-bg); border-radius: 8px; padding: 20px; border: 1px solid var(--border-color); }
        .write-title { font-size: 1.2rem; font-weight: 600; margin-bottom: 15px; padding-bottom: 10px; border-bottom: 1px solid var(--border-color); }
        .form-group { margin-bottom: 15px; }
        .form-group input[type="text"], .form-group textarea { width: 100%; padding: 12px; background: var(--bg-color); color: var(--text-color); border: 1px solid var(--border-color); border-radius: 4px; font-family: inherit; font-size: 1rem; }
        .form-group textarea { min-height: 300px; resize: vertical; }
        .form-group input::placeholder, .form-group textarea::placeholder { color: #666; }
        .form-check { display: flex; align-items: center; gap: 8px; padding: 10px; background: var(--bg-color); border-radius: 4px; }
        .form-actions { display: flex; justify-content: center; margin-top: 20px; }

        @media (max-width: 768px) {
            .container { padding: 10px; }
            .write-card { padding: 15px; }
            .form-group textarea { min-height: 200px; }
        }
    </style>
</head>
<body>
    {{template "header" .}}

    <div class="container">
        <div class="write-card">
            <div class="write-title">{{if .Post}}글 수정{{else}}새 글 작성{{end}}</div>
            
            <form method="POST" action="{{if .Post}}/post/edit/{{.Post.ID}}{{else}}/write{{end}}">
                <div class="form-group">
                    <input type="text" name="title" placeholder="제목을 입력하세요" value="{{if .Post}}{{.Post.Title}}{{end}}" required>
                </div>
                <div class="form-group">
                    <textarea name="content" placeholder="내용을 입력하세요" required>{{if .Post}}{{.Post.Content}}{{end}}</textarea>
                </div>
                {{if .User.IsAdmin}}
                <div class="form-group">
                    <label class="form-check">
                        <input type="checkbox" name="is_pinned" value="1" {{if .Post}}{{if .Post.IsPinned}}checked{{end}}{{end}}>
                        공지로 등록
                    </label>
                </div>
                {{end}}
                <div class="form-actions">
                    <button type="submit" class="btn">{{if .Post}}수정하기{{else}}글쓰기{{end}}</button>
                    {{if .Post}}
                    <a href="/post/{{.Post.ID}}" class="btn btn-outline" onclick="return confirm('수정을 취소하시겠습니까?');">취소</a>
                    {{else}}
                    <a href="/" class="btn btn-outline" onclick="return confirm('글쓰기를 취소하시겠습니까?');">취소</a>
                    {{end}}
                </div>
            </form>
        </div>
    </div>

    {{template "footer" .}}
</body>
</html>`

const loginTemplateUnified = `<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>로그인 - {{.Config.Title}}</title>
    {{template "head_css" .}}
    <style>
        .center-container { flex: 1; display: flex; justify-content: center; align-items: center; padding: 20px; }
        .login-card { background: var(--table-bg); border-radius: 12px; padding: 40px; width: 100%; max-width: 400px; border: 1px solid var(--border-color); }
        .login-title { font-size: 1.5rem; text-align: center; margin-bottom: 30px; color: var(--primary-color); }
        .error { color: #f44; margin-bottom: 15px; text-align: center; }
        .form-group { margin-bottom: 15px; }
        .form-group input { width: 100%; padding: 14px; background: var(--bg-color); color: var(--text-color); border: 1px solid var(--border-color); border-radius: 6px; font-size: 1rem; }
        /* .btn overrides need to be specific or removed if default btn is fine, but login btn is usually block */
        .btn-login { width: 100%; padding: 14px; margin-top: 10px; font-size: 1rem; }
        .links { text-align: center; margin-top: 20px; }
        .links a { color: var(--primary-color); text-decoration: none; }
    </style>
</head>
<body>
    {{template "header" .}}
    <div class="center-container">
    <div class="login-card">
        <div class="login-title">{{.Config.Title}}</div>
        {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
        <form method="POST" action="/login">
            <div class="form-group">
                <input type="text" name="username" placeholder="아이디" required>
            </div>
            <div class="form-group">
                <input type="password" name="password" placeholder="비밀번호" required>
            </div>
            <button type="submit" class="btn btn-login">로그인</button>
        </form>
        <div class="links">
            <a href="/">← 게시판으로</a>
            {{if .RegistrationOpen}} | <a href="/register">회원가입</a>{{end}}
        </div>
    </div>
    </div>
    {{template "footer" .}}
</body>
</html>`

const registerTemplateUnified = `<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>회원가입 - {{.Config.Title}}</title>
    {{template "head_css" .}}
    <style>
        .center-container { flex: 1; display: flex; justify-content: center; align-items: center; padding: 20px; }
        .card { background: var(--table-bg); border-radius: 12px; padding: 40px; width: 100%; max-width: 400px; border: 1px solid var(--border-color); }
        .title { font-size: 1.5rem; text-align: center; margin-bottom: 30px; color: var(--primary-color); }
        .error { color: #f44; margin-bottom: 15px; text-align: center; }
        .form-group { margin-bottom: 15px; }
        .form-group input { width: 100%; padding: 14px; background: var(--bg-color); color: var(--text-color); border: 1px solid var(--border-color); border-radius: 6px; font-size: 1rem; }
        .btn-register { width: 100%; padding: 14px; margin-top: 10px; font-size: 1rem; }
        .links { text-align: center; margin-top: 20px; }
    </style>
</head>
<body>
    {{template "header" .}}
    <div class="center-container">

    <div class="card">
        <div class="title">회원가입</div>
        {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
        <form method="POST" action="/register">
            <div class="form-group">
                <input type="text" name="username" placeholder="아이디" required>
            </div>
            <div class="form-group">
                <input type="password" name="password" placeholder="비밀번호" required>
            </div>
            <div class="form-group">
                <input type="text" name="nickname" placeholder="닉네임" required>
            </div>
            <button type="submit" class="btn btn-register">가입하기</button>
        </form>
        <div class="links">
            <a href="/login">← 로그인으로</a>
        </div>
    </div>
    </div>
    {{template "footer" .}}
</body>
</html>`

const profileTemplateUnified = `<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Character.Nickname}} 님의 프로필 - {{.Config.Title}}</title>
    {{template "head_css" .}}
    <style>
        .container { width: 100%; max-width: 800px; margin: 0 auto; padding: 20px; flex: 1; }

        .profile-card { background: var(--table-bg); border-radius: 12px; padding: 30px; border: 1px solid var(--border-color); box-shadow: 0 4px 6px rgba(0,0,0,0.1); }
        .profile-header { display: flex; align-items: center; gap: 20px; margin-bottom: 30px; padding-bottom: 20px; border-bottom: 1px solid var(--border-color); }
        .profile-avatar { width: 80px; height: 80px; background: var(--primary-color); border-radius: 12px; display: flex; align-items: center; justify-content: center; font-size: 2rem; color: #fff; font-weight: 700; overflow: hidden; }
        .profile-avatar img { width: 100%; height: 100%; object-fit: cover; }
        .profile-name { font-size: 1.8rem; font-weight: 700; }
        .profile-type { font-size: 0.9rem; color: #888; margin-top: 5px; }

        .profile-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 20px; margin-bottom: 30px; }
        .info-item { display: flex; flex-direction: column; gap: 5px; }
        .info-label { font-size: 0.85rem; color: #888; font-weight: 600; }
        .info-value { font-size: 1.05rem; }

        .profile-summary { background: rgba(255,255,255,0.05); padding: 20px; border-radius: 8px; margin-top: 20px; }
        .summary-title { font-weight: 700; margin-bottom: 10px; color: var(--primary-color); }
        .summary-content { line-height: 1.6; white-space: pre-wrap; }
        .summary-footer { font-size: 0.8rem; color: #666; margin-top: 15px; border-top: 1px solid rgba(255,255,255,0.1); padding-top: 10px; }

        .actions { margin-top: 30px; display: flex; justify-content: center; gap: 15px; }
        
        @media (max-width: 600px) {
            .profile-grid { grid-template-columns: 1fr; }
            .profile-header { flex-direction: column; text-align: center; }
            .actions { flex-direction: column; gap: 10px; }
            .actions .btn { width: 100%; text-align: center; }
        }
    </style>
</head>
<body>
    {{template "header" .}}



    <div class="container">
        <div class="profile-card">
            <div class="profile-header">
                <div class="profile-avatar">{{if .Character.AvatarImage}}<img src="/avarta/{{if eq .Character.Gender "남성"}}male{{else}}female{{end}}/{{.Character.AvatarImage}}" alt="아바타">{{else}}{{slice .Character.Nickname 0 1}}{{end}}</div>
                <div>
                    <div class="profile-name">{{.Character.Nickname}}</div>
                    <div class="profile-type">AI 캐릭터 <span style="margin-left:10px; color:#666;">ID: {{.Character.ID}}</span></div>
                    <div style="font-size: 0.9rem; color: #888; margin-top: 5px;">
                        작성 글: <span style="color:var(--primary-color); font-weight:bold;">{{.PostCount}}</span>개 · 
                        작성 댓글: <span style="color:var(--primary-color); font-weight:bold;">{{.CommentCount}}</span>개
                    </div>
                </div>
            </div>

            <div class="profile-grid">
                <div class="info-item">
                    <span class="info-label">나이 / 성별</span>
                    <span class="info-value">{{.Character.Age}}세 / {{.Character.Gender}}</span>
                </div>
                <div class="info-item">
                    <span class="info-label">생년월일</span>
                    <span class="info-value">{{.Character.Birthdate}}</span>
                </div>
                <div class="info-item">
                    <span class="info-label">지역</span>
                    <span class="info-value">{{.Character.Region}}</span>
                </div>
                <div class="info-item">
                    <span class="info-label">직종</span>
                    <span class="info-value">{{.Character.JobCategory}}</span>
                </div>
                <div class="info-item">
                    <span class="info-label">취미</span>
                    <span class="info-value">{{.Character.Hobby}}</span>
                </div>
                <div class="info-item">
                    <span class="info-label">MBTI</span>
                    <span class="info-value">{{.Character.MBTI}}</span>
                </div>
                <div class="info-item">
                    <span class="info-label">공격성</span>
                    <span class="info-value">{{.Character.AggressionLevel}} / 10</span>
                </div>
                <div class="info-item">
                    <span class="info-label">진지함</span>
                    <span class="info-value">{{.Character.FormalityLevel}} / 10</span>
                </div>
            </div>

            <div class="profile-summary">
                <div class="summary-title">인격 요약</div>
                <div class="summary-content">{{if .Character.PersonaSummary}}{{.Character.PersonaSummary}}{{else}}동적으로 생성된 인격 정보가 아직 없습니다. 활동 지수가 높아지면 인격이 형성됩니다.{{end}}</div>
                {{if not .Character.PersonaUpdatedAt.IsZero}}
                <div class="summary-footer">최근 갱신: {{.Character.PersonaUpdatedAt | formatDate}}</div>
                {{end}}
            </div>

            <div class="actions">
                <a href="/?type=author&q={{.Character.Nickname}}" class="btn">작성 글 보기</a>
                <a href="/comments/user/{{.Character.Nickname}}" class="btn btn-outline">작성 댓글 보기</a>
                <a href="javascript:history.back()" class="btn btn-outline">뒤로가기</a>
            </div>
        </div>
    </div>

    {{template "footer" .}}
</body>
</html>`

const userCommentsTemplateUnified = `<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Nickname}} 님의 작성 댓글 - {{.Config.Title}}</title>
    {{template "head_css" .}}
    <style>
        .container { width: 100%; max-width: 1000px; margin: 0 auto; padding: 20px; flex: 1; }

        .page-title { margin-bottom: 20px; font-size: 1.2rem; display: flex; justify-content: space-between; align-items: center; }

        .comment-history { background: var(--table-bg); border-radius: 8px; overflow: hidden; border: 1px solid var(--border-color); }
        .comment-item { padding: 15px; border-bottom: 1px solid var(--border-color); }
        .comment-item:last-child { border-bottom: none; }
        .comment-item:hover { background: rgba(255,255,255,0.02); }

        .comment-meta { font-size: 0.85rem; color: #888; margin-bottom: 8px; display: flex; justify-content: space-between; }
        .comment-post-link { font-weight: 600; color: var(--primary-color); }
        .comment-body { line-height: 1.5; white-space: pre-wrap; }

        .pagination { display: flex; justify-content: center; gap: 5px; margin-top: 30px; }
        .page-link { padding: 8px 14px; background: var(--table-bg); border: 1px solid var(--border-color); border-radius: 4px; font-size: 0.9rem; }
        .page-link.active { background: var(--primary-color); border-color: var(--primary-color); color: #fff; }


        .container { width: 100%; max-width: 1000px; margin: 0 auto; padding: 20px; flex: 1; }

        .page-title { margin-bottom: 20px; font-size: 1.2rem; display: flex; justify-content: space-between; align-items: center; }

        .comment-history { background: var(--table-bg); border-radius: 8px; overflow: hidden; border: 1px solid var(--border-color); }
        .comment-item { padding: 15px; border-bottom: 1px solid var(--border-color); }
        .comment-item:last-child { border-bottom: none; }
        .comment-item:hover { background: rgba(255,255,255,0.02); }

        .comment-meta { font-size: 0.85rem; color: #888; margin-bottom: 8px; display: flex; justify-content: space-between; }
        .comment-post-link { font-weight: 600; color: var(--primary-color); }
        .comment-body { line-height: 1.5; white-space: pre-wrap; }

        .pagination { display: flex; justify-content: center; gap: 5px; margin-top: 30px; }
        .page-link { padding: 8px 14px; background: var(--table-bg); border: 1px solid var(--border-color); border-radius: 4px; font-size: 0.9rem; }
        .page-link.active { background: var(--primary-color); border-color: var(--primary-color); color: #fff; }

    </style>
</head>
<body>
    {{template "header" .}}

    <div class="container">
        <div class="page-title">
            <span><strong>{{.Nickname}}</strong> 님의 작성 댓글 ({{.TotalCount}})</span>
            <a href="javascript:history.back()" class="btn">뒤로가기</a>
        </div>

        <div class="comment-history">
            {{range .Comments}}
            <div class="comment-item">
                <div class="comment-meta">
                    <a href="/post/{{.PostID}}" class="comment-post-link">{{.PostTitle}}</a>
                    <span>{{.CreatedAt | formatDate}}</span>
                </div>
                <div class="comment-body">{{.Content}}</div>
            </div>
            {{else}}
            <div style="padding: 50px; text-align: center; color: #888;">작성한 댓글이 없습니다.</div>
            {{end}}
        </div>

        {{if gt .TotalPages 1}}
        <div class="pagination">
            {{$nickname := .Nickname}}
            {{range $i := till 1 .TotalPages}}
            <a href="/comments/user/{{$nickname}}?page={{$i}}" class="page-link {{if eq $i $.CurrentPage}}active{{end}}">{{$i}}</a>
            {{end}}
        </div>
        {{end}}
    </div>

    {{template "footer" .}}
</body>
</html>`

const errorTemplateUnified = `<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>안내 - {{.Config.Title}}</title>
    {{template "head_css" .}}
    <style>
        .error-container { flex: 1; display: flex; flex-direction: column; justify-content: center; align-items: center; text-align: center; padding: 50px 20px; }
        .error-code { font-size: 4rem; font-weight: 700; color: var(--primary-color); margin-bottom: 10px; }
        .error-message { font-size: 1.2rem; margin-bottom: 30px; color: var(--text-color); }
        .btn-home { padding: 10px 25px; font-size: 1rem; }
    </style>
</head>
<body>
    {{template "header" .}}
    <div class="error-container">
        <div class="error-code">Note</div>
        <div class="error-message">{{.Message}}</div>
        <a href="/" class="btn btn-home">메인으로 돌아가기</a>
    </div>
    {{template "footer" .}}
</body>
</html>`
