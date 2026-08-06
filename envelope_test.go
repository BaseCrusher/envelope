package envelope

import "testing"

func TestConvert(t *testing.T) {
	for _, tc := range []struct {
		name    string
		environ []string
		want    string
		wantErr bool
	}{
		{
			name: "nested maps and scalar typing",
			environ: []string{
				"APP_LOGLEVEL=info",
				"APP_PROCESSES__DB__PATH=/bin/postgres",
				"APP_PROCESSES__DB__PORT=5432",
				"APP_PROCESSES__DB__ENABLED=true",
				"OTHER=ignored",
			},
			want: "LOGLEVEL: info\n" +
				"PROCESSES:\n" +
				"    DB:\n" +
				"        ENABLED: true\n" +
				"        PATH: /bin/postgres\n" +
				"        PORT: 5432\n",
		},
		{
			name:    "list of scalars, numeric order",
			environ: []string{"APP_ARGS__0=--verbose", "APP_ARGS__10=--last", "APP_ARGS__2=--limit"},
			want:    "ARGS:\n    - --verbose\n    - --limit\n    - --last\n",
		},
		{
			name:    "list of maps",
			environ: []string{"APP_LISTENERS__0__PORT=80", "APP_LISTENERS__1__PORT=443", "APP_LISTENERS__1__TLS=true"},
			want:    "LISTENERS:\n    - PORT: 80\n    - PORT: 443\n      TLS: true\n",
		},
		{
			name:    "nested lists",
			environ: []string{"APP_MATRIX__0__0=a", "APP_MATRIX__0__1=b", "APP_MATRIX__1__0=c"},
			want:    "MATRIX:\n    - - a\n      - b\n    - - c\n",
		},
		{
			name:    "sparse indexes are closed",
			environ: []string{"APP_ARGS__0=a", "APP_ARGS__7=b"},
			want:    "ARGS:\n    - a\n    - b\n",
		},
		{
			name:    "digits in a name are not an index",
			environ: []string{"APP_ADDRESS_LINE_1=Bahnhofstrasse", "APP_S3_BUCKET=data"},
			want:    "ADDRESS_LINE_1: Bahnhofstrasse\nS3_BUCKET: data\n",
		},
		{
			name:    "escaped segment is a numeric key",
			environ: []string{"APP_PORTS___80__TLS=true"},
			want:    "PORTS:\n    \"80\":\n        TLS: true\n",
		},
		{
			name:    "case is preserved in names and values",
			environ: []string{"APP_processes__Db__ENVIRONMENT__DATABASE_URL=postgres://Localhost/App"},
			want: "processes:\n" +
				"    Db:\n" +
				"        ENVIRONMENT:\n" +
				"            DATABASE_URL: postgres://Localhost/App\n",
		},
		{
			name: "empty collections, null, forced string",
			environ: []string{
				"APP_ARGS=[]",
				"APP_ENV={}",
				"APP_NOTE=~",
				"APP_PORT=\"8080\"",
				"APP_DESC=key: value",
				"APP_BLANK=",
			},
			want: "ARGS: []\n" +
				"BLANK: \"\"\n" +
				"DESC: 'key: value'\n" +
				"ENV: {}\n" +
				"NOTE: null\n" +
				"PORT: \"8080\"\n",
		},
		{
			name:    "connection string survives",
			environ: []string{"APP_CONN=Server=vServer034.m-s.ch;Password=pi+%emRZBr-Dp70!;Encrypt=false"},
			want:    "CONN: Server=vServer034.m-s.ch;Password=pi+%emRZBr-Dp70!;Encrypt=false\n",
		},
		{
			name:    "value and block on the same key",
			environ: []string{"APP_DB=x", "APP_DB__PORT=5432"},
			wantErr: true,
		},
		{
			name:    "index mixed with a name",
			environ: []string{"APP_ARGS__0=a", "APP_ARGS__NAME=b"},
			wantErr: true,
		},
		{
			name:    "empty segment",
			environ: []string{"APP_A____B=x"},
			wantErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Marshal(tc.environ, "APP_")
			if tc.wantErr {
				if err == nil {
					t.Fatalf("want error, got:\n%s", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tc.want {
				t.Errorf("got:\n%s\nwant:\n%s", got, tc.want)
			}
		})
	}
}
