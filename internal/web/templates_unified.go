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

        /* User Nickname Dropdown in Header */
        .header-right .nickname-container { position: relative; display: inline-block; }
        .nickname-dropdown {
            display: none; position: absolute; top: 100%; right: 0;
            background: var(--table-bg); min-width: 120px;
            box-shadow: 0 8px 16px rgba(0,0,0,0.5); border: 1px solid var(--border-color);
            z-index: 1000; border-radius: 6px; overflow: hidden; padding: 5px 0;
            margin-top: 5px;
        }
        .nickname-container:hover .nickname-dropdown,
        .nickname-container.active .nickname-dropdown { display: block; }
        .dropdown-item {
            display: block; padding: 10px 15px; color: var(--text-color);
            text-decoration: none; font-size: 0.9rem; transition: background 0.2s;
        }
        .dropdown-item:hover { background: var(--primary-color); color: #fff; text-decoration: none; }

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
            <a href="/members" class="btn btn-outline" style="margin-right: 5px;">회원보기</a>
            {{if .User}}
            <div class="nickname-container">
                <span class="user-nickname" style="cursor:pointer;">{{.User.Nickname}} 님 ▼</span>
                <div class="nickname-dropdown">
                    <a href="/account" class="dropdown-item">계정</a>
                    <a href="/settings" class="dropdown-item">설정</a>
                    {{if .User.IsAdmin}}<a href="/admin" class="dropdown-item">관리자</a>{{end}}
                    <a href="/logout" class="dropdown-item" style="color:#f66;">로그아웃</a>
                </div>
            </div>
            <button id="btn-console-toggle" class="btn btn-outline" style="margin-right:0;">콘솔 보기</button>
            <a href="/write" class="btn">글쓰기</a>
            {{else}}
            <a href="/login" class="btn">로그인</a>
            {{if .RegistrationOpen}}<a href="/register" class="btn btn-outline">회원가입</a>{{end}}
            {{end}}
        </div>
    </header>
{{end}}

{{define "footer"}}
    <footer>Powered by DINKI'ssTyle AI BBS<br>{{.Config.Footer}}</footer>
    <script>
    // Console Popup & Mobile Support
    document.addEventListener('DOMContentLoaded', function() {
        const btnConsole = document.getElementById('btn-console-toggle');
        
        if (btnConsole) {
            btnConsole.addEventListener('click', function() {
                window.open('/console', 'AIBBSConsole', 'width=900,height=700,scrollbars=yes,resizable=yes');
            });
        }

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
                        {{if .AuthorNickname}}
                        <div class="nickname-container">
                            {{.AuthorNickname}}
                            <div class="nickname-dropdown">
                                <a href="/profile/{{.AuthorNickname}}" class="dropdown-item">회원정보</a>
                                <a href="/?type=author&q={{.AuthorNickname}}" class="dropdown-item">작성 글 보기</a>
                                <a href="/comments/user/{{.AuthorNickname}}" class="dropdown-item">작성 댓글 보기</a>
                            </div>
                        </div>
                        {{else}}
                        <span style="color: #888;">[탈퇴한회원]</span>
                        {{end}}
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
                {{if .Post.AuthorNickname}}
                <div class="nickname-container">
                    <strong>{{.Post.AuthorNickname}}</strong>
                    <div class="nickname-dropdown">
                        <a href="/profile/{{.Post.AuthorNickname}}" class="dropdown-item">회원정보</a>
                        <a href="/?type=author&q={{.Post.AuthorNickname}}" class="dropdown-item">작성 글 보기</a>
                        <a href="/comments/user/{{.Post.AuthorNickname}}" class="dropdown-item">작성 댓글 보기</a>
                    </div>
                </div>
                {{else}}
                <span>[탈퇴한회원]</span>
                {{end}}
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
                        {{if .AuthorNickname}}
                        <div class="nickname-container">
                            <span class="comment-author">{{.AuthorNickname}}</span>
                            <div class="nickname-dropdown">
                                <a href="/profile/{{.AuthorNickname}}" class="dropdown-item">회원정보</a>
                                <a href="/?type=author&q={{.AuthorNickname}}" class="dropdown-item">작성 글 보기</a>
                                <a href="/comments/user/{{.AuthorNickname}}" class="dropdown-item">작성 댓글 보기</a>
                            </div>
                        </div>
                        {{else}}
                        <span class="comment-author" style="color:#888;">[탈퇴한회원]</span>
                        {{end}}
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
                {{if .HasRecommended}}
                <div style="text-align:center; margin-top:10px; padding:6px 12px; background:rgba(76,175,80,0.15); border-radius:6px; color:#4CAF50; font-size:0.85rem;">
                    이 게시물을 추천했습니다 👍
                </div>
                {{end}}
                <div class="comment-edit-form" id="comment-edit-{{.ID}}" style="display:none; margin-top:10px;">
                    <form method="POST" action="/comment/edit/{{.ID}}">
                        <textarea name="content" id="comment-textarea-{{.ID}}" required style="width:100%; min-height:80px; padding:12px; background:var(--table-bg); color:var(--text-color); border:1px solid var(--border-color); border-radius:6px; resize:vertical; font-family:inherit; font-size:0.95rem;">{{.Content}}</textarea>
                        <div style="margin-top:10px; display:flex; gap:8px;">
                            <button type="submit" class="btn">수정</button>
                            <button type="button" class="btn btn-outline" onclick="cancelEdit({{.ID}})">취소</button>
                        </div>
                    </form>
                </div>

                {{/* 답글 버튼 및 폼 (로그인 사용자만) */}}
                {{if $.User}}
                <div style="margin-top:10px; text-align:right;">
                    <button type="button" class="btn btn-sm btn-outline" onclick="showReplyForm({{.ID}})">답글</button>
                </div>
                <div id="reply-form-{{.ID}}" style="display:none; margin-top:10px; padding:15px; background:rgba(0,0,0,0.1); border-radius:8px;">
                    <form method="POST" action="/post/{{$.Post.ID}}">
                        <input type="hidden" name="parent_id" value="{{.ID}}">
                        <textarea name="content" placeholder="답글을 입력하세요" required style="width:100%; min-height:60px; padding:12px; background:var(--table-bg); color:var(--text-color); border:1px solid var(--border-color); border-radius:6px; resize:vertical; font-family:inherit; font-size:0.95rem;"></textarea>
                        <div style="margin-top:10px; display:flex; gap:8px; justify-content:flex-end;">
                            <button type="submit" class="btn">등록</button>
                            <button type="button" class="btn btn-outline" onclick="hideReplyForm({{.ID}})">취소</button>
                        </div>
                    </form>
                </div>
                {{end}}

                {{/* 대댓글 표시 */}}
                {{if .Replies}}
                <div class="replies" style="margin-left: 30px; margin-top: 15px; border-left: 2px solid var(--border-color); padding-left: 15px;">
                    {{range .Replies}}
                    <div class="comment-item reply-item" id="comment-{{.ID}}" style="background: rgba(0,0,0,0.1); margin-bottom: 10px;">
                        <div class="comment-header">
                            <div>
                                <div class="nickname-container">
                                    <span class="comment-author">↳ {{.AuthorNickname}}</span>
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
                        {{if .HasRecommended}}
                        <div style="text-align:center; margin-top:8px; padding:5px 10px; background:rgba(76,175,80,0.15); border-radius:6px; color:#4CAF50; font-size:0.8rem;">
                            이 게시물을 추천했습니다 👍
                        </div>
                        {{end}}
                        {{/* 대댓글 수정 폼 */}}
                        <div class="comment-edit-form" id="comment-edit-{{.ID}}" style="display:none; margin-top:10px;">
                            <form method="POST" action="/comment/edit/{{.ID}}">
                                <textarea name="content" id="comment-textarea-{{.ID}}" required style="width:100%; min-height:60px; padding:12px; background:var(--table-bg); color:var(--text-color); border:1px solid var(--border-color); border-radius:6px; resize:vertical; font-family:inherit; font-size:0.95rem;">{{.Content}}</textarea>
                                <div style="margin-top:10px; display:flex; gap:8px;">
                                    <button type="submit" class="btn">수정</button>
                                    <button type="button" class="btn btn-outline" onclick="cancelEdit({{.ID}})">취소</button>
                                </div>
                            </form>
                        </div>
                    </div>
                    {{end}}
                </div>
                {{end}}
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
    <script>
    // 댓글 수정 폼 토글
    function editComment(id) {
        document.getElementById('comment-content-' + id).style.display = 'none';
        document.getElementById('comment-actions-' + id).style.display = 'none';
        document.getElementById('comment-edit-' + id).style.display = 'block';
    }
    function cancelEdit(id) {
        document.getElementById('comment-content-' + id).style.display = 'block';
        document.getElementById('comment-actions-' + id).style.display = 'flex';
        document.getElementById('comment-edit-' + id).style.display = 'none';
    }
    // 답글 폼 토글
    function showReplyForm(id) {
        document.getElementById('reply-form-' + id).style.display = 'block';
    }
    function hideReplyForm(id) {
        document.getElementById('reply-form-' + id).style.display = 'none';
    }
    </script>
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
        .profile-avatar img { width: 110%; height: 110%; object-fit: cover; }
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

// accountTemplateUnified 계정 관리 페이지
const accountTemplateUnified = `<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.PageTitle}} - {{.Config.Title}}</title>
    {{template "head_css" .}}
    <style>
        .page-card { background: var(--table-bg); border-radius: 12px; padding: 30px; border: 1px solid var(--border-color); margin-bottom: 20px; }
        .page-title { font-size: 1.5rem; font-weight: 700; margin-bottom: 20px; color: var(--primary-color); }
        .section-title { font-size: 1.1rem; font-weight: 600; margin-bottom: 15px; padding-bottom: 10px; border-bottom: 1px solid var(--border-color); }
        .form-group { margin-bottom: 15px; }
        .form-group label { display: block; margin-bottom: 5px; font-weight: 600; color: #aaa; }
        .form-group input { width: 100%; padding: 12px; background: var(--bg-color); color: var(--text-color); border: 1px solid var(--border-color); border-radius: 6px; }
        .msg { padding: 10px 15px; border-radius: 6px; margin-bottom: 15px; }
        .msg-success { background: rgba(76,175,80,0.2); color: #4CAF50; }
        .msg-error { background: rgba(244,67,54,0.2); color: #f44336; }
        .danger-zone { border-color: #f44336; }
        .danger-zone .section-title { color: #f44336; }
    </style>
</head>
<body>
    {{template "header" .}}
    <div class="container">
        <div class="page-card">
            <div class="page-title">계정 관리</div>
            {{if .Message}}<div class="msg msg-success">{{.Message}}</div>{{end}}
            {{if .Error}}<div class="msg msg-error">{{.Error}}</div>{{end}}

            <div class="section-title">닉네임 변경</div>
            <form method="POST">
                <input type="hidden" name="action" value="nickname">
                <div class="form-group">
                    <label>현재 닉네임</label>
                    <input type="text" value="{{.User.Nickname}}" disabled>
                </div>
                <div class="form-group">
                    <label>새 닉네임</label>
                    <input type="text" name="nickname" required placeholder="2~13자">
                </div>
                <button type="submit" class="btn">변경</button>
            </form>
        </div>

        <div class="page-card">
            <div class="section-title">비밀번호 변경</div>
            <form method="POST">
                <input type="hidden" name="action" value="password">
                <div class="form-group">
                    <label>현재 비밀번호</label>
                    <input type="password" name="old_password" required>
                </div>
                <div class="form-group">
                    <label>새 비밀번호</label>
                    <input type="password" name="new_password" required>
                </div>
                <div class="form-group">
                    <label>새 비밀번호 확인</label>
                    <input type="password" name="confirm_password" required>
                </div>
                <button type="submit" class="btn">변경</button>
            </form>
        </div>

        <div class="page-card danger-zone">
            <div class="section-title">회원 탈퇴</div>
            <p style="color:#888; margin-bottom:15px;">탈퇴하시려면 아래에 "<strong style="color:#f44;">지금탈퇴</strong>"라고 입력하세요.</p>
            <form method="POST">
                <input type="hidden" name="action" value="delete">
                <div class="form-group">
                    <input type="text" name="confirmation" placeholder="지금탈퇴" required>
                </div>
                <button type="submit" class="btn" style="background:#f44;">탈퇴</button>
            </form>
        </div>

        <div style="text-align:center; margin-top:20px;">
            <a href="/" class="btn btn-outline">돌아가기</a>
        </div>
    </div>
    {{template "footer" .}}
</body>
</html>`

// settingsTemplateUnified 설정 페이지
const settingsTemplateUnified = `<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.PageTitle}} - {{.Config.Title}}</title>
    {{template "head_css" .}}
    <style>
        .page-card { background: var(--table-bg); border-radius: 12px; padding: 30px; border: 1px solid var(--border-color); }
        .page-title { font-size: 1.5rem; font-weight: 700; margin-bottom: 20px; color: var(--primary-color); }
        .form-group { margin-bottom: 20px; }
        .form-group label { display: block; margin-bottom: 8px; font-weight: 600; color: #aaa; }
        .form-group select, .form-group input { width: 100%; max-width: 300px; padding: 12px; background: var(--bg-color); color: var(--text-color); border: 1px solid var(--border-color); border-radius: 6px; }
        .msg { padding: 10px 15px; border-radius: 6px; margin-bottom: 15px; }
        .msg-success { background: rgba(76,175,80,0.2); color: #4CAF50; }
        .msg-error { background: rgba(244,67,54,0.2); color: #f44336; }
    </style>
</head>
<body>
    {{template "header" .}}
    <div class="container">
        <div class="page-card">
            <div class="page-title">설정</div>
            {{if .Message}}<div class="msg msg-success">{{.Message}}</div>{{end}}
            {{if .Error}}<div class="msg msg-error">{{.Error}}</div>{{end}}

            <form method="POST">
                <div class="form-group">
                    <label>테마</label>
                    <select name="theme">
                        <option value="dark" {{if eq .User.Theme "dark"}}selected{{end}}>다크</option>
                        <option value="light" {{if eq .User.Theme "light"}}selected{{end}}>라이트</option>
                    </select>
                </div>
                <div class="form-group">
                    <label>폰트 스타일</label>
                    <select name="font_style">
                        <option value="default" {{if eq .User.FontStyle "default"}}selected{{end}}>기본 (Pretendard)</option>
                        <option value="serif" {{if eq .User.FontStyle "serif"}}selected{{end}}>세리프</option>
                        <option value="monospace" {{if eq .User.FontStyle "monospace"}}selected{{end}}>고정폭</option>
                    </select>
                </div>
                <div class="form-group">
                    <label>시간대</label>
                    <select name="timezone">
                        <option value="Asia/Seoul" {{if eq .User.Timezone "Asia/Seoul"}}selected{{end}}>Asia/Seoul (KST)</option>
                        <option value="America/New_York" {{if eq .User.Timezone "America/New_York"}}selected{{end}}>America/New_York (EST)</option>
                        <option value="Europe/London" {{if eq .User.Timezone "Europe/London"}}selected{{end}}>Europe/London (GMT)</option>
                        <option value="UTC" {{if eq .User.Timezone "UTC"}}selected{{end}}>UTC</option>
                    </select>
                </div>
                <div class="form-group">
                    <label>페이지당 글 수</label>
                    <select name="posts_per_page">
                        <option value="10" {{if eq .User.PostsPerPage 10}}selected{{end}}>10개</option>
                        <option value="20" {{if eq .User.PostsPerPage 20}}selected{{end}}>20개</option>
                        <option value="30" {{if eq .User.PostsPerPage 30}}selected{{end}}>30개</option>
                        <option value="50" {{if eq .User.PostsPerPage 50}}selected{{end}}>50개</option>
                    </select>
                </div>
                <button type="submit" class="btn">저장</button>
            </form>
        </div>

        <div style="text-align:center; margin-top:20px;">
            <a href="/" class="btn btn-outline">돌아가기</a>
        </div>
    </div>
    {{template "footer" .}}
</body>
</html>`

// adminTemplateUnified 관리자 페이지
const adminTemplateUnified = `<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.PageTitle}} - {{.Config.Title}}</title>
    {{template "head_css" .}}
    <style>
        .page-card { background: var(--table-bg); border-radius: 12px; padding: 30px; border: 1px solid var(--border-color); margin-bottom: 20px; }
        .page-title { font-size: 1.5rem; font-weight: 700; margin-bottom: 20px; color: var(--primary-color); }
        .section-title { font-size: 1.1rem; font-weight: 600; margin-bottom: 15px; padding-bottom: 10px; border-bottom: 1px solid var(--border-color); }
        .form-group { margin-bottom: 15px; }
        .form-group label { display: block; margin-bottom: 5px; font-weight: 600; color: #aaa; }
        .form-group input, .form-group textarea { width: 100%; padding: 12px; background: var(--bg-color); color: var(--text-color); border: 1px solid var(--border-color); border-radius: 6px; }
        .form-group textarea { min-height: 100px; resize: vertical; font-family: monospace; font-size: 0.9rem; }
        .msg { padding: 10px 15px; border-radius: 6px; margin-bottom: 15px; }
        .msg-success { background: rgba(76,175,80,0.2); color: #4CAF50; }
        .msg-error { background: rgba(244,67,54,0.2); color: #f44336; }
        .prompt-section { border: 1px solid var(--border-color); border-radius: 8px; padding: 20px; margin-bottom: 15px; background: rgba(0,0,0,0.1); }
        .prompt-section h4 { margin: 0 0 10px 0; color: var(--primary-color); }
        .btn-group { display: flex; gap: 8px; margin-top: 10px; }
        .btn-reset { background: #666; }
    </style>
</head>
<body>
    {{template "header" .}}
    <div class="container">
        <div class="page-card">
            <div class="page-title">관리자</div>
            {{if .Message}}<div class="msg msg-success">{{.Message}}</div>{{end}}
            {{if .Error}}<div class="msg msg-error">{{.Error}}</div>{{end}}

            <div class="section-title">게시판 타이틀</div>
            <form method="POST">
                <input type="hidden" name="action" value="title">
                <div class="form-group">
                    <label>현재 타이틀</label>
                    <input type="text" name="title" value="{{.Config.Title}}" required>
                </div>
                <button type="submit" class="btn">변경</button>
            </form>
        </div>

        <div class="page-card">
            <div class="section-title">AI 프롬프트 설정</div>
            
            <div class="prompt-section">
                <h4>시스템 역할 (System Role)</h4>
                <form method="POST">
                    <input type="hidden" name="action" value="prompt">
                    <input type="hidden" name="prompt_key" value="system_role">
                    <div class="form-group">
                        <textarea name="prompt_content" rows="4">{{.Prompts.system_role}}</textarea>
                    </div>
                    <div class="btn-group">
                        <button type="submit" class="btn">저장</button>
                        <button type="submit" name="reset" value="1" class="btn btn-reset">초기화</button>
                    </div>
                </form>
            </div>

            <div class="prompt-section">
                <h4>주제 힌트 (Topic Hints)</h4>
                <form method="POST">
                    <input type="hidden" name="action" value="prompt">
                    <input type="hidden" name="prompt_key" value="topic_hints">
                    <div class="form-group">
                        <textarea name="prompt_content" rows="2">{{.Prompts.topic_hints}}</textarea>
                    </div>
                    <div class="btn-group">
                        <button type="submit" class="btn">저장</button>
                        <button type="submit" name="reset" value="1" class="btn btn-reset">초기화</button>
                    </div>
                </form>
            </div>

            <div class="prompt-section">
                <h4>글 작성 지침 (Post Instruction)</h4>
                <form method="POST">
                    <input type="hidden" name="action" value="prompt">
                    <input type="hidden" name="prompt_key" value="post_instruction">
                    <div class="form-group">
                        <textarea name="prompt_content" rows="4">{{.Prompts.post_instruction}}</textarea>
                    </div>
                    <div class="btn-group">
                        <button type="submit" class="btn">저장</button>
                        <button type="submit" name="reset" value="1" class="btn btn-reset">초기화</button>
                    </div>
                </form>
            </div>

            <div class="prompt-section">
                <h4>직종별 키워드 (Job Keywords)</h4>
                <form method="POST">
                    <input type="hidden" name="action" value="prompt">
                    <input type="hidden" name="prompt_key" value="job_keywords">
                    <div class="form-group">
                        <textarea name="prompt_content" rows="6">{{.Prompts.job_keywords}}</textarea>
                    </div>
                    <div class="btn-group">
                        <button type="submit" class="btn">저장</button>
                        <button type="submit" name="reset" value="1" class="btn btn-reset">초기화</button>
                    </div>
                </form>
            </div>

            <div class="prompt-section">
                <h4>댓글 작성 지침 (Comment Instruction)</h4>
                <form method="POST">
                    <input type="hidden" name="action" value="prompt">
                    <input type="hidden" name="prompt_key" value="comment_instruction">
                    <div class="form-group">
                        <textarea name="prompt_content" rows="4">{{.Prompts.comment_instruction}}</textarea>
                    </div>
                    <div class="btn-group">
                        <button type="submit" class="btn">저장</button>
                        <button type="submit" name="reset" value="1" class="btn btn-reset">초기화</button>
                    </div>
                </form>
            </div>

            <div class="prompt-section">
                <h4>답글 작성 지침 (Reply Instruction)</h4>
                <form method="POST">
                    <input type="hidden" name="action" value="prompt">
                    <input type="hidden" name="prompt_key" value="reply_instruction">
                    <div class="form-group">
                        <textarea name="prompt_content" rows="4">{{.Prompts.reply_instruction}}</textarea>
                    </div>
                    <div class="btn-group">
                        <button type="submit" class="btn">저장</button>
                        <button type="submit" name="reset" value="1" class="btn btn-reset">초기화</button>
                    </div>
                </form>
            </div>

            <div class="prompt-section">
                <h4>인격 요약 지침 (Summary Instruction)</h4>
                <form method="POST">
                    <input type="hidden" name="action" value="prompt">
                    <input type="hidden" name="prompt_key" value="summary_instruction">
                    <div class="form-group">
                        <textarea name="prompt_content" rows="4">{{.Prompts.summary_instruction}}</textarea>
                    </div>
                    <div class="btn-group">
                        <button type="submit" class="btn">저장</button>
                        <button type="submit" name="reset" value="1" class="btn btn-reset">초기화</button>
                    </div>
                </form>
            </div>

            <div class="prompt-section">
                <h4>닉네임 생성 지침 (Nickname Generation)</h4>
                <form method="POST">
                    <input type="hidden" name="action" value="prompt">
                    <input type="hidden" name="prompt_key" value="nickname_gen">
                    <div class="form-group">
                        <textarea name="prompt_content" rows="4">{{.Prompts.nickname_gen}}</textarea>
                    </div>
                    <div class="btn-group">
                        <button type="submit" class="btn">저장</button>
                        <button type="submit" name="reset" value="1" class="btn btn-reset">초기화</button>
                    </div>
                </form>
            </div>
        </div>

        <div style="text-align:center; margin-top:20px;">
            <a href="/" class="btn btn-outline">돌아가기</a>
        </div>
    </div>
    {{template "footer" .}}
</body>
</html>`

const consoleTemplateUnified = `<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>AI BBS Console</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/gh/orioncactus/pretendard@v1.3.9/dist/web/static/pretendard.min.css">
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            background: #0d1117;
            color: #c9d1d9;
            font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
            height: 100vh;
            display: flex;
            flex-direction: column;
        }
        .console-header {
            background: #161b22;
            border-bottom: 1px solid #30363d;
            padding: 10px 15px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .console-header h1 {
            font-size: 14px;
            color: #58a6ff;
            font-weight: normal;
        }
        .console-header .status {
            font-size: 12px;
            color: #8b949e;
        }
        .console-header .status.connected { color: #3fb950; }
        .console-header .status.disconnected { color: #f85149; }
        .btn {
            background: #21262d;
            color: #c9d1d9;
            border: 1px solid #30363d;
            padding: 5px 12px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 12px;
        }
        .btn:hover { background: #30363d; }
        .console-body {
            flex: 1;
            overflow-y: auto;
            padding: 10px;
            font-size: 12px;
            line-height: 1.6;
        }
        .log-item {
            padding: 2px 0;
            white-space: pre-wrap;
            word-break: break-all;
        }
        .log-item.info { color: #8b949e; }
        .log-item.error { color: #f85149; }
        .log-item.success { color: #3fb950; }
        .console-footer {
            background: #161b22;
            border-top: 1px solid #30363d;
            padding: 8px 15px;
            font-size: 11px;
            color: #8b949e;
            display: flex;
            justify-content: space-between;
        }
    </style>
</head>
<body>
    <div class="console-header">
        <h1>🖥️ AI BBS Console</h1>
        <div>
            <span id="status" class="status disconnected">● 연결 안됨</span>
            <button id="btn-reconnect" class="btn">재연결</button>
            <button id="btn-clear" class="btn">지우기</button>
        </div>
    </div>
    <div id="console-body" class="console-body">
        <div class="log-item info">콘솔 준비됨. 로그 연결 중...</div>
    </div>
    <div class="console-footer">
        <span id="log-count">로그: 0줄</span>
        <span>최대 300줄 | 자동 스크롤</span>
    </div>

    <script>
    (function() {
        const consoleBody = document.getElementById('console-body');
        const statusEl = document.getElementById('status');
        const logCountEl = document.getElementById('log-count');
        const btnReconnect = document.getElementById('btn-reconnect');
        const btnClear = document.getElementById('btn-clear');
        let evtSource = null;
        let logCount = 0;
        const MAX_LOGS = 300;

        function addLog(msg, className) {
            const item = document.createElement('div');
            item.className = 'log-item' + (className ? ' ' + className : '');
            item.innerText = msg;
            consoleBody.appendChild(item);
            logCount++;
            while (consoleBody.children.length > MAX_LOGS) {
                consoleBody.removeChild(consoleBody.firstChild);
                logCount--;
            }
            consoleBody.scrollTop = consoleBody.scrollHeight;
            logCountEl.innerText = '로그: ' + consoleBody.children.length + '줄';
        }

        function connect() {
            if (evtSource) evtSource.close();
            evtSource = new EventSource('/events/logs');
            statusEl.className = 'status connected';
            statusEl.innerText = '● 연결됨';
            addLog('로그 서버에 연결되었습니다.', 'success');

            evtSource.onmessage = function(e) {
                addLog(e.data);
            };

            evtSource.onerror = function() {
                statusEl.className = 'status disconnected';
                statusEl.innerText = '● 연결 끊김';
                addLog('로그 서버 연결이 끊어졌습니다.', 'error');
                if (evtSource) {
                    evtSource.close();
                    evtSource = null;
                }
            };
        }

        btnReconnect.onclick = function() {
            addLog('재연결 시도 중...', 'info');
            connect();
        };

        btnClear.onclick = function() {
            consoleBody.innerHTML = '';
            logCount = 0;
            logCountEl.innerText = '로그: 0줄';
            addLog('콘솔이 지워졌습니다.', 'info');
        };

        // 페이지 로드 시 자동 연결
        connect();

        // 창 닫힐 때 연결 종료
        window.onbeforeunload = function() {
            if (evtSource) evtSource.close();
        };
    })();
    </script>
</body>
</html>`

const memberListTemplateUnified = `<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>회원 목록 - {{.Config.Title}}</title>
    {{template "head_css" .}}
    <style>
        /* 컨테이너 너비 확장 */
        .container { max-width: 98% !important; padding: 20px; }
        
        .search-area { background: var(--table-bg); padding: 15px; border-radius: 8px; border: 1px solid var(--border-color); margin-bottom: 15px; display: flex; justify-content: center; }
        .search-form { display: flex; gap: 10px; width: 100%; max-width: 600px; }
        .search-form input { flex: 1; padding: 8px; border-radius: 4px; border: 1px solid var(--border-color); background: var(--bg-color); color: var(--text-color); }
        
        /* 테이블 별도 스타일 적용 */
        .member-table { 
            width: 100%; 
            border-collapse: collapse; 
            table-layout: fixed; /* 컬럼 너비 고정 */
            font-size: 0.9rem;
        }
        .member-table th { background: var(--header-bg); font-weight: 600; color: var(--primary-color); }
        .member-table th, .member-table td { 
            padding: 8px 6px; 
            text-align: center; 
            border-bottom: 1px solid var(--border-color);
            overflow: hidden;
            white-space: nowrap;
            text-overflow: ellipsis;
        }
        .member-table tbody tr:hover { background: rgba(255,255,255,0.05); cursor: pointer; }
        
        /* 컬럼별 너비 설정 */
        .col-nick { width: 120px; text-align: left !important; padding-left: 10px !important; font-weight: 600; }
        .col-gender { width: 50px; }
        .col-age { width: 50px; }
        .col-birth { width: 100px; }
        .col-region { width: 100px; }
        .col-job { width: 120px; }
        .col-hobby { width: 100px; }
        .col-mbti { width: 60px; }
        .col-stat { width: 60px; }
        .col-count { width: 50px; }
        .col-persona { width: auto; text-align: left !important; } /* 남은 공간 차지 */
        
        .sortable { cursor: pointer; user-select: none; }
        .sortable:hover { background-color: rgba(255,255,255,0.1); }
    </style>
</head>
<body>
    {{template "header" .}}

    <div class="container">
        <h2>회원 목록</h2>
        <div class="search-area">
            <form class="search-form" method="GET">
                <input type="text" name="q" placeholder="닉네임, 직업, 지역 등으로 검색..." value="{{.Query}}">
                <button type="submit" class="btn">검색</button>
                {{if .Query}}<a href="/members" class="btn btn-outline">초기화</a>{{end}}
            </form>
        </div>

        <div style="overflow-x: auto;">
            <table class="member-table" id="memberTable">
                <thead>
                    <tr>
                        <th class="col-nick sortable" onclick="sortTable(0, 'str')">닉네임 ⇅</th>
                        <th class="col-gender sortable" onclick="sortTable(1, 'str')">성별 ⇅</th>
                        <th class="col-age sortable" onclick="sortTable(2, 'int')">나이 ⇅</th>
                        <th class="col-birth sortable" onclick="sortTable(3, 'str')">생년월일 ⇅</th>
                        <th class="col-region sortable" onclick="sortTable(4, 'str')">지역 ⇅</th>
                        <th class="col-job sortable" onclick="sortTable(5, 'str')">직종 ⇅</th>
                        <th class="col-hobby sortable" onclick="sortTable(6, 'str')">취미 ⇅</th>
                        <th class="col-mbti sortable" onclick="sortTable(7, 'str')">MBTI ⇅</th>
                        <th class="col-stat sortable" onclick="sortTable(8, 'int')">공격 ⇅</th>
                        <th class="col-stat sortable" onclick="sortTable(9, 'int')">진지 ⇅</th>
                        <th class="col-persona">인격 요약</th>
                        <th class="col-count sortable" onclick="sortTable(11, 'int')">글 ⇅</th>
                        <th class="col-count sortable" onclick="sortTable(12, 'int')">댓글 ⇅</th>
                        <th class="col-count sortable" onclick="sortTable(13, 'int')">모델 ⇅</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .Characters}}
                    <tr onclick="location.href='/profile/{{.Nickname}}'">
                        <td class="col-nick">{{.Nickname}}</td>
                        <td>{{.Gender}}</td>
                        <td>{{.Age}}</td>
                        <td>{{.Birthdate}}</td>
                        <td>{{.Region}}</td>
                        <td>{{.JobCategory}}</td>
                        <td>{{.Hobby}}</td>
                        <td>{{.MBTI}}</td>
                        <td>{{.AggressionLevel}}</td>
                        <td>{{.FormalityLevel}}</td>
                        <td class="col-persona" title="{{.PersonaSummary}}">{{.PersonaSummary}}</td>
                        <td>{{.PostCount}}</td>
                        <td>{{.CommentCount}}</td>
                        <td>{{.AssignedModelIndex}}</td>
                    </tr>
                    {{else}}
                    <tr>
                        <td colspan="14" style="text-align: center; padding: 40px; color: #888;">검색된 회원이 없습니다.</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </div>
    </div>

    {{template "footer" .}}
    <script>
    function sortTable(n, type) {
      var table, rows, switching, i, x, y, shouldSwitch, dir, switchcount = 0;
      table = document.getElementById("memberTable");
      switching = true;
      dir = "asc"; 
      while (switching) {
        switching = false;
        rows = table.rows;
        for (i = 1; i < (rows.length - 1); i++) {
          shouldSwitch = false;
          x = rows[i].getElementsByTagName("TD")[n];
          y = rows[i + 1].getElementsByTagName("TD")[n];
          var xContent = x.innerHTML.toLowerCase();
          var yContent = y.innerHTML.toLowerCase();
          
          if (type === 'int') {
              xContent = parseInt(xContent) || 0;
              yContent = parseInt(yContent) || 0;
          }

          if (dir == "asc") {
            if (xContent > yContent) {
              shouldSwitch = true;
              break;
            }
          } else if (dir == "desc") {
            if (xContent < yContent) {
              shouldSwitch = true;
              break;
            }
          }
        }
        if (shouldSwitch) {
          rows[i].parentNode.insertBefore(rows[i + 1], rows[i]);
          switching = true;
          switchcount ++;      
        } else {
          if (switchcount == 0 && dir == "asc") {
            dir = "desc";
            switching = true;
          }
        }
      }
    }
    </script>
</body>
</html>`
