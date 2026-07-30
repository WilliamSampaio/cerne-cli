//go:build windows

package filecheck

import (
	"errors"

	"golang.org/x/sys/windows"
)

func probe(path string, read bool) (Outcome, string) {
	access := uint32(windows.GENERIC_WRITE)
	if read {
		access = windows.GENERIC_READ
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
