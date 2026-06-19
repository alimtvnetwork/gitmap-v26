import DocsLayout from "@/components/docs/DocsLayout";
import CodeBlock from "@/components/docs/CodeBlock";
import { Chrome, Download, Upload, Database } from "lucide-react";

const commands = [
  { name: "chrome-profile-copy", alias: "cpc", desc: "Copy a Chrome profile directory (bookmarks, extensions, prefs, flags). Emits JSON + CSV snapshots and upserts the SQLite ChromeProfile table." },
  { name: "chrome-profile-export", alias: "cpe", desc: "Export a named profile to JSON + a sibling CSV. Both paths are printed and recorded in ChromeProfileExport." },
  { name: "chrome-profile-import", alias: "cpi", desc: "Restore a profile from a JSON snapshot." },
  { name: "chrome-profile-list", alias: "cpl", desc: "List profiles discovered on disk + every profile tracked in the gitmap database (with export counts and last-seen timestamps)." },
];

const ChromeProfileSpec = () => (
  <DocsLayout>
    <div className="max-w-4xl space-y-10">
      <div>
        <div className="flex items-center gap-3 mb-2">
          <Chrome className="h-8 w-8 text-primary" />
          <h1 className="text-3xl font-bold tracking-tight">Chrome profile management</h1>
        </div>
        <p className="text-lg text-muted-foreground">
          Offline copy, export, import, and audit of Google Chrome profiles. Each
          export writes both JSON and CSV, and persists into the local SQLite DB so
          you can list every profile gitmap has touched.
        </p>
      </div>

      <section>
        <h2 className="text-xl font-semibold mb-3">Commands</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {commands.map((c) => (
            <div key={c.name} className="rounded-lg border border-border p-4 bg-card">
              <div className="flex items-center gap-2 mb-1">
                <code className="font-mono text-sm text-primary">{c.name}</code>
                <span className="font-mono text-xs px-2 py-0.5 rounded bg-primary/10 text-foreground border border-primary/20">
                  {c.alias}
                </span>
              </div>
              <p className="text-xs text-muted-foreground">{c.desc}</p>
            </div>
          ))}
        </div>
      </section>

      <section>
        <h2 className="text-xl font-semibold mb-3 flex items-center gap-2">
          <Download className="h-5 w-5 text-primary" /> Export — JSON + CSV
        </h2>
        <CodeBlock code={`# Default output: .gitmap/chrome/<name>.json plus .gitmap/chrome/<name>.csv
gitmap chrome-profile-export Default
gitmap cpe "Profile 1" ./snapshots/work.json   # CSV sibling: ./snapshots/work.csv`} />
        <p className="text-sm text-muted-foreground mt-2">
          Both file paths are printed on stdout and inserted as separate rows in
          the <code>ChromeProfileExport</code> table.
        </p>
      </section>

      <section>
        <h2 className="text-xl font-semibold mb-3 flex items-center gap-2">
          <Upload className="h-5 w-5 text-primary" /> Import
        </h2>
        <CodeBlock code={`gitmap chrome-profile-import ./snapshots/work.json
gitmap cpi ./snapshots/work.json "Profile 2"   # optional destination profile name`} />
      </section>

      <section>
        <h2 className="text-xl font-semibold mb-3 flex items-center gap-2">
          <Database className="h-5 w-5 text-primary" /> SQLite tables
        </h2>
        <CodeBlock code={`ChromeProfile (
  ChromeProfileId  INTEGER PRIMARY KEY AUTOINCREMENT,
  Name             TEXT UNIQUE,
  SourcePath       TEXT,
  IsOffline        INTEGER,
  CreatedAt        TEXT,
  UpdatedAt        TEXT
)

ChromeProfileExport (
  ChromeProfileExportId  INTEGER PRIMARY KEY AUTOINCREMENT,
  ChromeProfileId        INTEGER REFERENCES ChromeProfile,
  Format                 TEXT,   -- 'json' | 'csv'
  FilePath               TEXT,
  ByteSize               INTEGER,
  CreatedAt              TEXT
)`} />
      </section>

      <section>
        <h2 className="text-xl font-semibold mb-3">List both disk + DB</h2>
        <CodeBlock code={`gitmap chrome-profile-list
# Chrome profiles (/.../Chrome/User Data):
#   - Default
#   - Profile 1
# Tracked in gitmap DB:
#   - Default        exports=4  last=2026-06-19T08:14:02Z
#   - Profile 1      exports=1  last=2026-06-19T08:14:09Z`} />
      </section>
    </div>
  </DocsLayout>
);

export default ChromeProfileSpec;
