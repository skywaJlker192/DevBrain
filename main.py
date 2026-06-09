# main.py
from fastapi import FastAPI, HTTPException, Request
from fastapi.staticfiles import StaticFiles
from fastapi.responses import HTMLResponse
import sqlite3, httpx
from bs4 import BeautifulSoup
from contextlib import asynccontextmanager

app = FastAPI(title="DevBrain")
DB = "devbrain.db"

def init_db():
    with sqlite3.connect(DB) as conn:
        conn.execute("""
            CREATE TABLE IF NOT EXISTS links (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                url TEXT UNIQUE NOT NULL,
                title TEXT,
                description TEXT,
                tags TEXT,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            )
        """)
init_db()

def fetch_meta(url: str):
    try:
        with httpx.Client(timeout=5, follow_redirects=True) as c:
            r = c.get(url, headers={"User-Agent": "DevBrain/1.0"})
            r.raise_for_status()
            soup = BeautifulSoup(r.text, "html.parser")
            title = soup.title.string.strip() if soup.title else ""
            desc = ""
            meta = soup.find("meta", attrs={"name": "description"})
            if meta: desc = meta.get("content", "").strip()
            return {"title": title, "desc": desc}
    except Exception:
        return {"title": "", "desc": ""}

@app.post("/api/links")
async def add_link(req: Request):
    data = await req.json()
    url, tags = data.get("url", "").strip(), data.get("tags", "").strip()
    if not url or not url.startswith(("http://", "https://")):
        raise HTTPException(400, "Укажите корректный URL")

    meta = fetch_meta(url)
    try:
        with sqlite3.connect(DB) as conn:
            conn.execute("INSERT INTO links (url, title, description, tags) VALUES (?,?,?,?)",
                         (url, meta["title"], meta["desc"], tags))
        return {"status": "ok"}
    except sqlite3.IntegrityError:
        return {"status": "exists"}

@app.get("/api/links")
def get_links(q: str = "", tags: str = ""):
    with sqlite3.connect(DB) as conn:
        cur = conn.cursor()
        query, params = "SELECT * FROM links WHERE 1=1 ORDER BY created_at DESC", []
        if q:
            query += " AND (title LIKE ? OR description LIKE ? OR url LIKE ?)"
            params.extend([f"%{q}%", f"%{q}%", f"%{q}%"])
        if tags:
            for t in [t.strip() for t in tags.split(",") if t.strip()]:
                query += " AND tags LIKE ?"
                params.append(f"%{t}%")
        cur.execute(query, params)
        return cur.fetchall()

@app.delete("/api/links/{lid}")
def delete_link(lid: int):
    with sqlite3.connect(DB) as conn:
        conn.execute("DELETE FROM links WHERE id=?", (lid,))
    return {"status": "deleted"}

app.mount("/static", StaticFiles(directory="static"), name="static")

@app.get("/", response_class=HTMLResponse)
async def index():
    with open("static/index.html", encoding="utf-8") as f: return f.read()
