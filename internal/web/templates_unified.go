// Created by DINKIssTyle on 2025. Copyright (C) 2025 DINKI'ssTyle. All rights reserved.

package web

// Unified Responsive Template - 모던웹/모바일 통합 반응형 템플릿

const boardTemplateUnified = `<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.PageTitle}} - {{.Config.Title}}</title>
    <link href="https://fonts.googleapis.com/css2?family=Pretendard:wght@400;600;700&display=swap" rel="stylesheet">
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

        .container { width: 100%; max-width: 1000px; margin: 0 auto; padding: 20px; flex: 1; }

        .board-table { width: 100%; border-collapse: collapse; background: var(--table-bg); border-radius: 8px; overflow: hidden; }
        .board-table thead { background: var(--header-bg); }
        .board-table th, .board-table td { padding: 8px 10px; text-align: center; border-bottom: 1px solid var(--border-color); height: 50px; vertical-align: middle; }
        .board-table th { font-weight: 600; color: var(--primary-color); }
        .board-table td.title { text-align: left; }
        .board-table tr:hover { background: rgba(255,255,255,0.05); }
        .board-table .pinned { background: rgba(255, 215, 0, 0.1); }

        .col-id { width: 60px; }
        .col-author { width: 120px; }
        .col-date { width: 100px; font-size: 0.85rem; line-height: 1.2; }
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

        footer { padding: 20px; text-align: center; color: #888; font-size: 0.85rem; border-top: 1px solid var(--border-color); }

        @media (max-width: 768px) {
            .header { flex-direction: column; align-items: flex-start; }
            .header h1 { font-size: 1.2rem; }
            .container { padding: 10px; }
            .board-table thead { display: none; }
            .board-table, .board-table tbody, .board-table tr, .board-table td { display: block; width: 100%; }
            .board-table tr { background: var(--table-bg); margin-bottom: 10px; padding: 12px; border-radius: 8px; border: 1px solid var(--border-color); }
            .board-table td { padding: 4px 0; text-align: left; border: none; }
            .board-table td.title { font-weight: 600; font-size: 1rem; margin-bottom: 8px; }
            .board-table td.meta { font-size: 0.8rem; color: #888; }
            .board-table td:before { content: attr(data-label); font-weight: 600; color: var(--primary-color); margin-right: 8px; }
            .controls-bar { flex-direction: column; align-items: stretch; }
            .search-form { width: 100%; }
            .search-form input[type="text"] { flex: 1; }
            .controls-left, .controls-right { width: 100%; justify-content: space-between; }
        }
    </style>
</head>
<body>
    <header class="header">
        <div class="header-left">
            <a href="/"><h1>{{.Config.Title}}</h1></a>
        </div>
        <div class="header-right">
            {{if .User}}
            <span>{{.User.Nickname}} 님</span>
            <a href="/write" class="btn">글쓰기</a>
            <a href="/logout" class="btn btn-outline">로그아웃</a>
            {{else}}
            <a href="/login" class="btn">로그인</a>
            {{if .RegistrationOpen}}<a href="/register" class="btn btn-outline">회원가입</a>{{end}}
            {{end}}
        </div>
    </header>

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
                    <td class="col-author" data-label="작성자: ">{{.AuthorNickname}}</td>
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

    <footer>{{.Config.Footer}}</footer>
</body>
</html>`

const postTemplateUnified = `<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Post.Title}} - {{.Config.Title}}</title>
    <link href="https://fonts.googleapis.com/css2?family=Pretendard:wght@400;600;700&display=swap" rel="stylesheet">
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
        body { background: var(--bg-color); color: var(--text-color); font-family: 'Pretendard', -apple-system, BlinkMacSystemFont, sans-serif; min-height: 100vh; display: flex; flex-direction: column; }
        a { color: var(--link-color); text-decoration: none; }
        a:hover { text-decoration: underline; }

        .header { background: var(--header-bg); border-bottom: 2px solid var(--primary-color); padding: 15px 20px; display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 10px; }
        .header-left { display: flex; align-items: center; gap: 15px; }
        .header h1 { font-size: 1.5rem; color: var(--primary-color); }
        .header-right { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }

        .btn { background: var(--primary-color); color: #fff; border: none; padding: 8px 16px; border-radius: 4px; cursor: pointer; font-weight: 600; font-size: 0.9rem; transition: opacity 0.2s; }
        .btn:hover { opacity: 0.85; text-decoration: none; }
        .btn-outline { background: transparent; border: 1px solid var(--primary-color); color: var(--primary-color); }
        .btn-sm { padding: 4px 10px; font-size: 0.8rem; }

        .container { width: 100%; max-width: 1000px; margin: 0 auto; padding: 20px; flex: 1; }

        .post-card { background: var(--table-bg); border-radius: 8px; padding: 20px; border: 1px solid var(--border-color); }
        .post-title { font-size: 1.4rem; font-weight: 700; margin-bottom: 10px; padding-bottom: 10px; border-bottom: 1px solid var(--border-color); }
        .post-meta { font-size: 0.85rem; color: #888; margin-bottom: 15px; }
        .post-content { line-height: 1.7; min-height: 200px; white-space: pre-wrap; }
        .post-actions { display: flex; justify-content: center; gap: 10px; margin-top: 25px; padding-top: 15px; border-top: 1px solid var(--border-color); }

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
            .header { flex-direction: column; align-items: flex-start; }
            .container { padding: 10px; }
            .post-title { font-size: 1.2rem; }
            .comment-form { flex-direction: column; }
            .comment-form .btn { width: 100%; }
        }
    </style>
</head>
<body>
    <header class="header">
        <div class="header-left">
            <a href="/"><h1>{{.Config.Title}}</h1></a>
        </div>
        <div class="header-right">
            {{if .User}}
            <span>{{.User.Nickname}} 님</span>
            <a href="/write" class="btn">글쓰기</a>
            <a href="/logout" class="btn btn-outline">로그아웃</a>
            {{else}}
            <a href="/login" class="btn">로그인</a>
            {{end}}
        </div>
    </header>

    <div class="container">
        <div class="post-card">
            <div class="post-title">{{if .Post.IsPinned}}<span style="color:var(--primary-color);">[공지]</span> {{end}}{{.Post.Title}}</div>
            <div class="post-meta">
                <strong>{{.Post.AuthorNickname}}</strong> · {{.Post.CreatedAt | formatDate}} · 조회 {{.Post.ViewCount}} · 추천 {{.Post.RecommendCount}}
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
                        <span class="comment-author">{{.AuthorNickname}}</span>
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

    <footer>{{.Config.Footer}}</footer>
    <script>
    function editComment(id) {
        document.getElementById('comment-content-' + id).style.display = 'none';
        var actions = document.getElementById('comment-actions-' + id);
        if (actions) actions.style.display = 'none';
        document.getElementById('comment-edit-' + id).style.display = 'block';
        document.getElementById('comment-textarea-' + id).focus();
    }
    function cancelEdit(id) {
        document.getElementById('comment-content-' + id).style.display = 'block';
        var actions = document.getElementById('comment-actions-' + id);
        if (actions) actions.style.display = 'flex';
        document.getElementById('comment-edit-' + id).style.display = 'none';
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
    <link href="https://fonts.googleapis.com/css2?family=Pretendard:wght@400;600;700&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg-color: {{.ResultColors.BgColor}};
            --text-color: {{.ResultColors.TextColor}};
            --primary-color: {{.ResultColors.PointColor}};
            --table-bg: {{.ResultColors.TableBgColor}};
            --header-bg: {{.ResultColors.HeaderBgColor}};
            --border-color: {{.ResultColors.BorderColor}};
        }

        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { background: var(--bg-color); color: var(--text-color); font-family: 'Pretendard', -apple-system, BlinkMacSystemFont, sans-serif; min-height: 100vh; display: flex; flex-direction: column; }
        a { color: var(--primary-color); text-decoration: none; }

        .header { background: var(--header-bg); border-bottom: 2px solid var(--primary-color); padding: 15px 20px; display: flex; justify-content: space-between; align-items: center; }
        .header-left { display: flex; align-items: center; gap: 15px; }
        .header h1 { font-size: 1.5rem; color: var(--primary-color); }

        .btn { background: var(--primary-color); color: #fff; border: none; padding: 8px 16px; border-radius: 4px; cursor: pointer; font-weight: 600; font-size: 0.9rem; }
        .btn:hover { opacity: 0.85; }
        .btn-outline { background: transparent; border: 1px solid var(--primary-color); color: var(--primary-color); }

        .container { width: 100%; max-width: 1000px; margin: 0 auto; padding: 20px; flex: 1; }

        .write-card { background: var(--table-bg); border-radius: 8px; padding: 20px; border: 1px solid var(--border-color); }
        .write-title { font-size: 1.2rem; font-weight: 600; margin-bottom: 15px; padding-bottom: 10px; border-bottom: 1px solid var(--border-color); }
        .form-group { margin-bottom: 15px; }
        .form-group input[type="text"], .form-group textarea { width: 100%; padding: 12px; background: var(--bg-color); color: var(--text-color); border: 1px solid var(--border-color); border-radius: 4px; font-family: inherit; font-size: 1rem; }
        .form-group textarea { min-height: 300px; resize: vertical; }
        .form-group input::placeholder, .form-group textarea::placeholder { color: #666; }
        .form-check { display: flex; align-items: center; gap: 8px; padding: 10px; background: var(--bg-color); border-radius: 4px; }
        .form-actions { display: flex; justify-content: center; margin-top: 20px; }

        footer { padding: 20px; text-align: center; color: #888; font-size: 0.85rem; border-top: 1px solid var(--border-color); }

        @media (max-width: 768px) {
            .container { padding: 10px; }
            .write-card { padding: 15px; }
            .form-group textarea { min-height: 200px; }
        }
    </style>
</head>
<body>
    <header class="header">
        <div class="header-left">
            <a href="/"><h1>{{.Config.Title}}</h1></a>
        </div>
        <div class="header-right">
            {{if .User}}
            <span>{{.User.Nickname}} 님</span>
            <a href="/logout" class="btn btn-outline">로그아웃</a>
            {{else}}
            <a href="/login" class="btn">로그인</a>
            {{end}}
        </div>
    </header>

    <div class="container">
        <div class="write-card">
            <div class="write-title">{{if .Post}}글 수정{{else}}새 글 작성{{end}}</div>
            
            <form method="POST" action="{{if .Post}}/edit/{{.Post.ID}}{{else}}/write{{end}}">
                <div class="form-group">
                    <input type="text" name="title" placeholder="제목을 입력하세요" value="{{if .Post}}{{.Post.Title}}{{end}}" required>
                </div>
                <div class="form-group">
                    <textarea name="content" placeholder="내용을 입력하세요" required>{{if .Post}}{{.Post.Content}}{{end}}</textarea>
                </div>
                {{if .IsAdmin}}
                <div class="form-group">
                    <label class="form-check">
                        <input type="checkbox" name="is_pinned" {{if .Post}}{{if .Post.IsPinned}}checked{{end}}{{end}}>
                        공지로 등록
                    </label>
                </div>
                {{end}}
                <div class="form-actions">
                    <button type="submit" class="btn">{{if .Post}}수정하기{{else}}글쓰기{{end}}</button>
                    <button type="button" class="btn btn-outline" onclick="if(confirm('정말 글쓰기를 취소하시겠습니까?')) history.back();">취소</button>
                </div>
            </form>
        </div>
    </div>

    <footer>{{.Config.Footer}}</footer>
</body>
</html>`

const loginTemplateUnified = `<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>로그인 - {{.Config.Title}}</title>
    <style>
        :root {
            --bg-color: {{.ResultColors.BgColor}};
            --text-color: {{.ResultColors.TextColor}};
            --primary-color: {{.ResultColors.PointColor}};
            --table-bg: {{.ResultColors.TableBgColor}};
            --border-color: {{.ResultColors.BorderColor}};
        }

        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { background: var(--bg-color); color: var(--text-color); font-family: -apple-system, BlinkMacSystemFont, sans-serif; min-height: 100vh; display: flex; justify-content: center; align-items: center; padding: 20px; }

        .login-card { background: var(--table-bg); border-radius: 12px; padding: 40px; width: 100%; max-width: 400px; border: 1px solid var(--border-color); }
        .login-title { font-size: 1.5rem; text-align: center; margin-bottom: 30px; color: var(--primary-color); }
        .error { color: #f44; margin-bottom: 15px; text-align: center; }
        .form-group { margin-bottom: 15px; }
        .form-group input { width: 100%; padding: 14px; background: var(--bg-color); color: var(--text-color); border: 1px solid var(--border-color); border-radius: 6px; font-size: 1rem; }
        .btn { width: 100%; padding: 14px; background: var(--primary-color); color: #fff; border: none; border-radius: 6px; cursor: pointer; font-weight: 600; font-size: 1rem; margin-top: 10px; }
        .btn:hover { opacity: 0.9; }
        .links { text-align: center; margin-top: 20px; }
        .links a { color: var(--primary-color); text-decoration: none; }
    </style>
</head>
<body>
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
            <button type="submit" class="btn">로그인</button>
        </form>
        <div class="links">
            <a href="/">← 게시판으로</a>
            {{if .RegistrationOpen}} | <a href="/register">회원가입</a>{{end}}
        </div>
    </div>
</body>
</html>`

const registerTemplateUnified = `<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>회원가입 - {{.Config.Title}}</title>
    <style>
        :root {
            --bg-color: {{.ResultColors.BgColor}};
            --text-color: {{.ResultColors.TextColor}};
            --primary-color: {{.ResultColors.PointColor}};
            --table-bg: {{.ResultColors.TableBgColor}};
            --border-color: {{.ResultColors.BorderColor}};
        }

        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { background: var(--bg-color); color: var(--text-color); font-family: -apple-system, BlinkMacSystemFont, sans-serif; min-height: 100vh; display: flex; justify-content: center; align-items: center; padding: 20px; }

        .card { background: var(--table-bg); border-radius: 12px; padding: 40px; width: 100%; max-width: 400px; border: 1px solid var(--border-color); }
        .title { font-size: 1.5rem; text-align: center; margin-bottom: 30px; color: var(--primary-color); }
        .error { color: #f44; margin-bottom: 15px; text-align: center; }
        .form-group { margin-bottom: 15px; }
        .form-group input { width: 100%; padding: 14px; background: var(--bg-color); color: var(--text-color); border: 1px solid var(--border-color); border-radius: 6px; font-size: 1rem; }
        .btn { width: 100%; padding: 14px; background: var(--primary-color); color: #fff; border: none; border-radius: 6px; cursor: pointer; font-weight: 600; font-size: 1rem; margin-top: 10px; }
        .links { text-align: center; margin-top: 20px; }
        .links a { color: var(--primary-color); }
    </style>
</head>
<body>
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
            <button type="submit" class="btn">가입하기</button>
        </form>
        <div class="links">
            <a href="/login">← 로그인으로</a>
        </div>
    </div>
</body>
</html>`
