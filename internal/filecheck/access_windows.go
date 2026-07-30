//go:build windows

package filecheck

import (
	"errors"
	"os"

	"golang.org/x/sys/windows"
)

func probe(path string, read bool) (Outcome, string) {
	info, statErr := os.Lstat(path)
	if statErr != nil {
		return Unknown, "não foi possível confirmar permissões"
	}
	access := uint32(windows.FILE_WRITE_DATA)
	if info.IsDir() {
		access = windows.FILE_WRITE_DATA | windows.FILE_APPEND_DATA
	}
	if read && info.IsDir() {
		access = windows.FILE_LIST_DIRECTORY | windows.FILE_TRAVERSE | windows.FILE_READ_ATTRIBUTES
	} else if read {
		access = windows.FILE_READ_DATA | windows.FILE_READ_ATTRIBUTES
	}
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return Unknown, "não foi possível confirmar permissões"
	}
	handle, err := windows.CreateFile(name, access,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err == nil {
		windows.CloseHandle(handle)
		return Allowed, ""
	}
	if errors.Is(err, windows.ERROR_ACCESS_DENIED) {
		return Denied, "acesso negado"
	}
	return Unknown, "não foi possível confirmar permissões"
}
