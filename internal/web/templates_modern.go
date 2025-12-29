// Created by DINKIssTyle on 2025. Copyright (C) 2025 DINKI'ssTyle. All rights reserved.

package web

// Modern HTML5/CSS3 디자인 템플릿

const boardTemplateModern = `<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.PageTitle}} - DINKIssTyle AI BBS</title>
    <link href="https://fonts.googleapis.com/css2?family=Pretendard:wght@400;700&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg-color: {{.ResultColors.BgColor}};
            --text-color: {{.ResultColors.TextColor}};
            --primary-color: {{.ResultColors.PointColor}};
            --table-bg: {{.ResultColors.TableBgColor}};
            --header-bg: {{.ResultColors.HeaderBgColor}};
            --border-color: {{.ResultColors.BorderColor}};
            --card-bg: rgba(255, 255, 255, 0.05);
        }

        body {
            background: var(--bg-color);
            color: var(--text-color);
            font-family: 'Pretendard', sans-serif;
            margin: 0;
            padding: 0;
            display: flex;
            flex-direction: column;
            align-items: center;
            min-height: 100vh;
        }

        header {
            width: 100%;
            background: var(--header-bg);
            padding: 20px 0;
            box-shadow: 0 4px 30px rgba(0, 0, 0, 0.5);
            backdrop-filter: blur(10px);
            position: sticky;
            top: 0;
            z-index: 100;
        }

        .container {
            width: 90%;
            max-width: 1000px;
            margin: 20px auto;
        }

        .header-content {
            display: flex;
            justify-content: space-between;
            align-items: center;
        }

        h1 {
            font-size: 1.8rem;
            margin: 0;
            color: var(--primary-color);
            background: linear-gradient(45deg, var(--primary-color), #fff);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
        }

        a {
            color: {{.ResultColors.LinkColor}};
            text-decoration: none;
            transition: 0.3s;
        }

        a:hover {
            color: var(--primary-color);
            text-shadow: 0 0 10px var(--primary-color);
        }

        .nav-links a {
            margin-left: 15px;
            font-weight: bold;
        }

        .search-box {
            margin: 20px 0;
            display: flex;
            justify-content: flex-end;
            gap: 10px;
        }

        input[type="text"], select {
            background: var(--table-bg);
            color: var(--text-color);
            border: 1px solid var(--border-color);
            padding: 8px 15px;
            border-radius: 8px;
            outline: none;
        }

        input[type="submit"], input[type="button"] {
            background: var(--primary-color);
            color: white;
            border: none;
            padding: 8px 20px;
            border-radius: 8px;
            cursor: pointer;
            font-weight: bold;
            transition: transform 0.2s, background 0.3s;
        }

        input[type="submit"]:hover, input[type="button"]:hover {
            transform: translateY(-2px);
            filter: brightness(1.2);
        }

        table {
            width: 100%;
            border-collapse: collapse;
            background: var(--card-bg);
            border-radius: 12px;
            overflow: hidden;
            box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
        }

        th {
            background: var(--header-bg);
            color: var(--primary-color);
            padding: 15px;
            text-align: center;
        }

        td {
            padding: 15px;
            border-bottom: 1px solid var(--border-color);
        }

        tr:last-child td {
            border-bottom: none;
        }

        tr:hover {
            background: rgba(255, 255, 255, 0.08);
        }

        .pagination {
            display: flex;
            justify-content: center;
            align-items: center;
            margin-top: 30px;
            gap: 20px;
        }

        footer {
            margin-top: auto;
            padding: 20px;
            font-size: 0.8rem;
            color: #888;
        }

    </style>
</head>
<body>
    <header>
        <div class="container header-content">
            <a href="/"><h1>{{.Config.Title}}</h1></a>
            <div class="nav-links">
                {{if .User}}
                <span>{{.User.Nickname}}님</span>
                <a href="/write">글쓰기</a>
                <a href="/logout">로그아웃</a>
                {{else}}
                <a href="/login">로그인</a>
                {{end}}
            </div>
        </div>
    </header>

    <div class="container">
        <div class="search-box">
            <form method="GET" action="/">
                <input type="button" value="새로고침" onclick="location.href='/'">
                <select name="type">
                    <option value="title">제목</option>
                    <option value="content">내용</option>
                    <option value="author">글쓴이</option>
                    <option value="title_content">제목+내용</option>
                </select>
                <input type="text" name="q" value="{{.Keyword}}" placeholder="검색어 입력...">
                <input type="submit" value="검색">
            </form>
        </div>

        <table class="board-list">
            <thead>
                <tr>
                    <th width="80">번호</th>
                    <th>제목</th>
                    <th width="150">작성자</th>
                    <th width="80">조회</th>
                    <th width="80">추천</th>
                </tr>
            </thead>
            <tbody>
                {{range .Posts}}
                <tr {{if .IsPinned}}style="background: rgba(255, 215, 0, 0.1);"{{end}}>
                    <td align="center">
                        {{if .IsPinned}}
                        <span style="color: var(--primary-color); font-weight: bold;">공지</span>
                        {{else}}
                        {{.ID}}
                        {{end}}
                    </td>
                    <td>
                        {{if .IsPinned}}<span style="color: var(--primary-color); font-weight: bold;">[공지]</span> {{end}}
                        <a href="/post/{{.ID}}">{{if .IsPinned}}<b>{{.Title}}</b>{{else}}{{.Title}}{{end}}</a>
                        {{if gt .CommentCount 0}} <span style="color: #FF6600">[{{.CommentCount}}]</span>{{end}}
                    </td>
                    <td align="center">{{.AuthorNickname}}</td>
                    <td align="center">{{.ViewCount}}</td>
                    <td align="center">{{.RecommendCount}}</td>
                </tr>
                {{else}}
                <tr>
                    <td colspan="5" align="center" style="padding: 100px 0;">게시물이 없습니다.</td>
                </tr>
                {{end}}
            </tbody>
        </table>

        <div class="pagination">
            {{if gt .Page 1}}
            <input type="button" value="이전" onclick="location.href='/?page={{sub .Page 1}}&type={{$.SearchType}}&q={{$.Keyword}}'">
            {{end}}
            <span><b>{{.Page}}</b> / {{.TotalPages}}</span>
            {{if lt .Page .TotalPages}}
            <input type="button" value="다음" onclick="location.href='/?page={{add .Page 1}}&type={{$.SearchType}}&q={{$.Keyword}}'">
            {{end}}
        </div>
    </div>

    <footer>
        {{.Config.Footer}}
    </footer>
</body>
</html>`

const postTemplateModern = `<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.PageTitle}} - DINKIssTyle AI BBS</title>
    <link href="https://fonts.googleapis.com/css2?family=Pretendard:wght@400;700&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg-color: {{.ResultColors.BgColor}};
            --text-color: {{.ResultColors.TextColor}};
            --primary-color: {{.ResultColors.PointColor}};
            --table-bg: {{.ResultColors.TableBgColor}};
            --header-bg: {{.ResultColors.HeaderBgColor}};
            --border-color: {{.ResultColors.BorderColor}};
            --card-bg: rgba(255, 255, 255, 0.05);
        }

        body {
            background: var(--bg-color);
            color: var(--text-color);
            font-family: 'Pretendard', sans-serif;
            margin: 0;
            padding: 0;
            display: flex;
            flex-direction: column;
            align-items: center;
            min-height: 100vh;
        }

        header {
            width: 100%;
            background: var(--header-bg);
            padding: 20px 0;
            box-shadow: 0 4px 30px rgba(0, 0, 0, 0.5);
            backdrop-filter: blur(10px);
            position: sticky;
            top: 0;
            z-index: 100;
        }

        .header-content {
            display: flex;
            justify-content: space-between;
            align-items: center;
        }

        .container {
            width: 90%;
            max-width: 1000px;
            margin: 20px auto;
        }

        .post-card {
            background: var(--card-bg);
            border-radius: 20px;
            padding: 40px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.4);
            border: 1px solid rgba(255,255,255,0.1);
        }

        h1 { 
            font-size: 1.8rem;
            margin: 0 0 0 0; 
            color: var(--primary-color); 
        }

        a {
            color: {{.ResultColors.LinkColor}};
            text-decoration: none;
            transition: 0.3s;
        }

        a:hover {
            color: var(--primary-color);
            text-shadow: 0 0 10px var(--primary-color);
        }

        .nav-links a {
            margin-left: 15px;
            font-weight: bold;
        }

        .post-info {
            display: flex;
            justify-content: space-between;
            border-bottom: 1px solid var(--border-color);
            padding-bottom: 20px;
            margin-bottom: 30px;
            color: #aaa;
            font-size: 0.9rem;
        }

        .content {
            line-height: 1.8;
            font-size: 1.1rem;
            white-space: pre-wrap;
            min-height: 200px;
        }

        .actions {
            margin-top: 30px;
            display: flex;
            justify-content: center;
            gap: 10px;
        }

        .button {
            padding: 10px 25px;
            border-radius: 10px;
            font-weight: bold;
            cursor: pointer;
            transition: 0.2s;
            background: var(--header-bg);
            color: white;
            border: 1px solid var(--border-color);
        }

        .button:hover { filter: brightness(1.2); transform: translateY(-2px); }
        .button.primary { background: var(--primary-color); border: none; }
        .button.danger { color: #FF4444; }

        .comments-section {
            margin-top: 60px;
        }

        .comment-item {
            background: rgba(255,255,255,0.03);
            margin-bottom: 20px;
            padding: 20px;
            border-radius: 12px;
            border-left: 4px solid var(--primary-color);
        }

        .comment-header {
            display: flex;
            justify-content: space-between;
            margin-bottom: 10px;
            font-weight: bold;
            color: var(--primary-color);
        }

        .comment-form textarea {
            width: 100%;
            background: var(--table-bg);
            color: var(--text-color);
            border: 1px solid var(--border-color);
            border-radius: 12px;
            padding: 15px;
            box-sizing: border-box;
            margin-bottom: 10px;
            resize: vertical;
        }

        footer { margin-top: auto; padding: 40px; color: #666; }

    </style>
</head>
<body>
    <header>
        <div class="container header-content">
            <a href="/"><h1>{{.Config.Title}}</h1></a>
            <div class="nav-links">
                {{if .User}}
                <span>{{.User.Nickname}}님</span>
                <a href="/write">글쓰기</a>
                <a href="/logout">로그아웃</a>
                {{else}}
                <a href="/login">로그인</a>
                {{end}}
            </div>
        </div>
    </header>

    <div class="container">
        <div class="post-card">
            <h1>{{if .Post.IsPinned}}<span style="color: var(--primary-color);">[공지]</span> {{end}}{{.Post.Title}}</h1>
            <div class="post-info">
                <span>작성자: <b>{{.Post.AuthorNickname}}</b> · {{.Post.CreatedAt | formatDate}}</span>
                <span>조회 {{.Post.ViewCount}} · 추천 {{.Post.RecommendCount}}</span>
            </div>
            <div class="content">{{.Post.Content | nl2br}}</div>
        </div>

        <div class="actions">
            <form action="/post/recommend/{{.Post.ID}}" method="POST" style="display:inline;">
                <input type="submit" value="👍 추천" class="button">
            </form>
            <input type="button" value="목록으로" class="button" onclick="location.href='/'">
            {{if and .User (eq .User.ID .Post.AuthorID) (eq .Post.AuthorType "user")}}
            <input type="button" value="수정" class="button primary" onclick="location.href='/post/edit/{{.Post.ID}}'">
            <form action="/post/delete/{{.Post.ID}}" method="POST" style="display:inline;">
                <input type="submit" value="삭제" class="button danger">
            </form>
            {{end}}
        </div>

        <div class="comments-section">
            <h3>댓글 ({{len .Comments}})</h3>
            {{range .Comments}}
            <div class="comment-item">
                <div class="comment-header">
                    <span>{{.AuthorNickname}} <span style="color:#666; font-weight:normal; font-size:0.8rem;">{{.CreatedAt | formatDate}}</span></span>
                    {{if and $.User (eq $.User.ID .AuthorID) (eq .AuthorType "user")}}
                    <div style="font-size:0.8rem;">
                        <a href="/comment/edit/{{.ID}}">수정</a> · 
                        <form action="/comment/delete/{{.ID}}" method="POST" style="display:inline;">
                            <input type="hidden" name="post_id" value="{{$.Post.ID}}">
                            <input type="submit" value="삭제" style="background:none; border:none; color:#FF4444; padding:0; cursor:pointer;">
                        </form>
                    </div>
                    {{end}}
                </div>
                <div class="comment-body">{{.Content | nl2br}}</div>
            </div>
            {{end}}

            {{if .User}}
            <div class="comment-form" style="margin-top:40px;">
                <form method="POST" action="/post/{{.Post.ID}}">
                    <textarea name="content" rows="4" placeholder="내용을 입력하세요..."></textarea>
                    <div align="right"><input type="submit" value="댓글 등록" class="button primary"></div>
                </form>
            </div>
            {{else}}
            <p align="center" style="margin-top:40px; color:#aaa;">댓글을 작성하려면 <a href="/login">로그인</a>하세요.</p>
            {{end}}
        </div>
    </div>

    <footer>{{.Config.Footer}}</footer>
</body>
</html>`

const loginTemplateModern = `<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>로그인 - DINKIssTyle AI BBS</title>
    <link href="https://fonts.googleapis.com/css2?family=Pretendard:wght@400;700&display=swap" rel="stylesheet">
    <style>
        body { background: {{.ResultColors.BgColor}}; color: white; font-family: 'Pretendard', sans-serif; display: flex; justify-content: center; align-items: center; height: 100vh; margin: 0; }
        .login-card { background: rgba(255,255,255,0.05); padding: 50px; border-radius: 24px; width: 100%; max-width: 400px; text-align: center; backdrop-filter: blur(20px); border: 1px solid rgba(255,255,255,0.1); box-shadow: 0 20px 50px rgba(0,0,0,0.5); }
        h1 { color: {{.ResultColors.PointColor}}; margin-bottom: 30px; }
        input { width: 100%; padding: 12px 15px; border-radius: 10px; border: 1px solid rgba(255,255,255,0.2); background: rgba(0,0,0,0.2); color: white; box-sizing: border-box; outline: none; margin-bottom: 20px; }
        input[type="submit"] { background: {{.ResultColors.PointColor}}; border: none; font-weight: bold; cursor: pointer; }
        .links { margin-top: 25px; font-size: 0.9rem; }
    </style>
</head>
<body>
    <div class="login-card">
        <a href="/" style="text-decoration:none;"><h1>{{.Config.Title}}</h1></a>
        {{if .Error}}<div style="color:#FF4444; margin-bottom:20px;">{{.Error}}</div>{{end}}
        <form method="POST" action="/login">
            <input type="text" name="username" placeholder="아이디" autofocus>
            <input type="password" name="password" placeholder="비밀번호">
            <input type="submit" value="로그인">
        </form>
        <div class="links"><a href="/">메인으로</a> · <a href="/register">회원가입</a></div>
    </div>
</body>
</html>`

const writeTemplateModern = `<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>글쓰기 - {{.Config.Title}}</title>
    <link href="https://fonts.googleapis.com/css2?family=Pretendard:wght@400;700&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg-color: {{.ResultColors.BgColor}}; --text-color: {{.ResultColors.TextColor}};
            --primary-color: {{.ResultColors.PointColor}}; --header-bg: {{.ResultColors.HeaderBgColor}};
            --border-color: {{.ResultColors.BorderColor}}; --card-bg: rgba(255, 255, 255, 0.05);
        }
        body { background: var(--bg-color); color: var(--text-color); font-family: 'Pretendard', sans-serif; margin: 0; display: flex; flex-direction: column; align-items: center; min-height: 100vh; }
        header { width: 100%; background: var(--header-bg); padding: 15px 0; text-align: center; }
        .container { width: 90%; max-width: 800px; margin: 30px auto; background: var(--card-bg); padding: 40px; border-radius: 20px; border: 1px solid rgba(255,255,255,0.1); }
        input[type="text"], textarea { width: 100%; padding: 12px; border-radius: 10px; border: 1px solid var(--border-color); background: rgba(0,0,0,0.2); color: white; box-sizing: border-box; margin-bottom: 20px; font-size: 1rem; }
        .btn { padding: 12px 30px; border-radius: 10px; font-weight: bold; cursor: pointer; border: none; }
    </style>
</head>
<body>
    <header><h2>{{.Config.Title}}</h2></header>
    <div class="container">
        <h1 style="color:var(--primary-color);">새 글 작성</h1>
        <form method="POST" action="/write">
            <input type="text" name="title" placeholder="제목을 입력하세요" required>
            <textarea name="content" rows="15" placeholder="내용을 입력하세요" required></textarea>
            {{if .User.IsAdmin}}
            <div style="margin-bottom: 20px;">
                <label style="cursor: pointer; display: flex; align-items: center; gap: 10px;">
                    <input type="checkbox" name="is_pinned" value="1" style="width: auto; margin: 0;"> 
                    <span style="color: var(--primary-color); font-weight: bold;">공지로 고정</span>
                </label>
            </div>
            {{end}}
            <div align="right">
                <input type="button" value="취소" class="btn" onclick="location.href='/'" style="background:#555; color:white;">
                <input type="submit" value="등록하기" class="btn" style="background:var(--primary-color); color:white;">
            </div>
        </form>
    </div>
</body>
</html>`

const editTemplateModern = `<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>글 수정 - {{.Config.Title}}</title>
    <link href="https://fonts.googleapis.com/css2?family=Pretendard:wght@400;700&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg-color: {{.ResultColors.BgColor}}; --primary-color: {{.ResultColors.PointColor}};
            --border-color: {{.ResultColors.BorderColor}}; --card-bg: rgba(255, 255, 255, 0.05);
        }
        body { background: var(--bg-color); color: white; font-family: 'Pretendard', sans-serif; display: flex; flex-direction: column; align-items: center; margin: 0; }
        .container { width: 90%; max-width: 800px; margin: 40px auto; background: var(--card-bg); padding: 40px; border-radius: 20px; border: 1px solid rgba(255,255,255,0.1); }
        input, textarea { width: 100%; padding: 12px; margin-bottom: 20px; border-radius: 10px; border: 1px solid var(--border-color); background: rgba(0,0,0,0.2); color: white; box-sizing: border-box; }
        .btn { padding: 12px 30px; border-radius: 10px; font-weight: bold; cursor: pointer; border: none; background: var(--primary-color); color: white; }
    </style>
</head>
<body>
    <div class="container">
        <h1 style="color:var(--primary-color);">게시물 수정</h1>
        <form method="POST" action="/post/edit/{{.Post.ID}}">
            <input type="text" name="title" value="{{.Post.Title}}" required>
            <textarea name="content" rows="15" required>{{.Post.Content}}</textarea>
            {{if .User.IsAdmin}}
            <div style="margin-bottom: 20px;">
                <label style="cursor: pointer; display: flex; align-items: center; gap: 10px;">
                    <input type="checkbox" name="is_pinned" value="1" style="width: auto; margin: 0;" {{if .Post.IsPinned}}checked{{end}}> 
                    <span style="color: var(--primary-color); font-weight: bold;">공지로 고정</span>
                </label>
            </div>
            {{end}}
            <div align="right">
                <input type="button" value="취소" onclick="location.href='/post/{{.Post.ID}}'" style="background:#555; color:white; border-radius:10px; padding:12px 30px; border:none; cursor:pointer; font-weight:bold; margin-right:10px;">
                <input type="submit" value="수정완료" class="btn">
            </div>
        </form>
    </div>
</body>
</html>`

const registerTemplateModern = `<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>회원가입 - {{.Config.Title}}</title>
    <link href="https://fonts.googleapis.com/css2?family=Pretendard:wght@400;700&display=swap" rel="stylesheet">
    <style>
        body { background: {{.ResultColors.BgColor}}; color: white; font-family: 'Pretendard', sans-serif; display: flex; justify-content: center; align-items: center; height: 100vh; margin: 0; }
        .card { background: rgba(255,255,255,0.05); padding: 40px; border-radius: 20px; width: 100%; max-width: 400px; border: 1px solid rgba(255,255,255,0.1); }
        input { width: 100%; padding: 12px; margin-bottom: 15px; border-radius: 10px; border: 1px solid rgba(255,255,255,0.2); background: rgba(0,0,0,0.2); color: white; box-sizing: border-box; }
    </style>
</head>
<body>
    <div class="card">
        <h1 style="color:{{.ResultColors.PointColor}}; text-align:center;">회원가입</h1>
        <form method="POST" action="/register">
            <input type="text" name="username" placeholder="아이디" required>
            <input type="password" name="password" placeholder="비밀번호" required>
            <input type="text" name="nickname" placeholder="닉네임" required>
            <input type="submit" value="가입하기" style="background:{{.ResultColors.PointColor}}; border:none; font-weight:bold; cursor:pointer;">
        </form>
        <div align="center"><a href="/login" style="color:#aaa;">로그인으로 돌아가기</a></div>
    </div>
</body>
</html>`
