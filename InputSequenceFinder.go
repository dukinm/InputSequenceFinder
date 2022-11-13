package InputSequenceFinder

import (
	"github.com/hallazzang/go-windows-programming/pkg/win"
	"golang.org/x/sys/windows"
	"strings"
	"syscall"
	"unsafe"
)

type (
	DWORD     uint32
	WPARAM    uintptr
	LPARAM    uintptr
	LRESULT   uintptr
	HANDLE    uintptr
	HINSTANCE HANDLE
	HHOOK     HANDLE
	HWND      HANDLE
)

type HOOKPROC func(int, WPARAM, LPARAM) LRESULT

type KBDLLHOOKSTRUCT struct {
	VkCode      DWORD
	ScanCode    DWORD
	Flags       DWORD
	Time        DWORD
	DwExtraInfo uintptr
}

var (
	user32                  = windows.NewLazySystemDLL("user32.dll")
	procSetWindowsHookEx    = user32.NewProc("SetWindowsHookExW")
	procCallNextHookEx      = user32.NewProc("CallNextHookEx")
	procUnhookWindowsHookEx = user32.NewProc("UnhookWindowsHookEx")
	procGetMessage          = user32.NewProc("GetMessageW")
	procTranslateMessage    = user32.NewProc("TranslateMessage")
	procDispatchMessage     = user32.NewProc("DispatchMessageW")
	keyboardHook            HHOOK
)

const (
	WH_KEYBOARD_LL = 13
	WH_KEYBOARD    = 2
	WM_KEYDOWN     = 256
	WM_SYSKEYDOWN  = 260
	WM_KEYUP       = 257
	WM_SYSKEYUP    = 261
	WM_KEYFIRST    = 256
	WM_KEYLAST     = 264
	PM_NOREMOVE    = 0x000
	PM_REMOVE      = 0x001
	PM_NOYIELD     = 0x002
	WM_LBUTTONDOWN = 513
	WM_RBUTTONDOWN = 516
	NULL           = 0
)

func SetWindowsHookEx(idHook int, lpfn HOOKPROC, hMod HINSTANCE, dwThreadId DWORD) HHOOK {
	ret, _, _ := procSetWindowsHookEx.Call(
		uintptr(idHook),
		uintptr(syscall.NewCallback(lpfn)),
		uintptr(hMod),
		uintptr(dwThreadId),
	)
	return HHOOK(ret)
}

func CallNextHookEx(hhk HHOOK, nCode int, wParam WPARAM, lParam LPARAM) LRESULT {
	ret, _, _ := procCallNextHookEx.Call(
		uintptr(hhk),
		uintptr(nCode),
		uintptr(wParam),
		uintptr(lParam),
	)
	return LRESULT(ret)
}

type UpdateFunc func(input string)

func Detect(startToken string, endToken string, callback UpdateFunc) {
	line := ""
	shiftWasDown := false

	keyboardHook = SetWindowsHookEx(WH_KEYBOARD_LL,
		func(nCode int, wparam WPARAM, lparam LPARAM) LRESULT {
			if nCode == 0 && wparam == WM_KEYUP {
				kbdstruct := (*KBDLLHOOKSTRUCT)(unsafe.Pointer(lparam))
				if kbdstruct.VkCode == 160 {
					shiftWasDown = false
				}
			} else if nCode == 0 && wparam == WM_KEYDOWN {
				kbdstruct := (*KBDLLHOOKSTRUCT)(unsafe.Pointer(lparam))
				code := string(rune(kbdstruct.VkCode))

				if kbdstruct.VkCode == 160 { //
					shiftWasDown = true
				}
				if code != " " && code != "	" && code != " " && code != "\b" {
					if code == "Û" {
						code = "{"
					} else if code == "Þ" {
						code = "\""
					} else if code == "º" {
						code = ":"
					} else if code == "¼" {
						code = ","
					} else if code == "Ý" {
						code = "}"
					} else if code == "¾" {
						code = "."
					} else if code == "½" {
						code = "_"
					} else if code == "2" && shiftWasDown {
						code = "@"
					}
					line += code
					shiftWasDown = false
					if strings.Contains(line, startToken) && strings.Contains(line, endToken) {
						startCommandIndex := strings.Index(line, startToken)
						endCommandIndex := strings.Index(line, endToken)
						command := line[(startCommandIndex + len(startToken)):endCommandIndex]
						line = ""
						callback(command)
					}
				}
			}
			return CallNextHookEx(keyboardHook, nCode, wparam, lparam)
		}, 0, 0)

	var msg win.MSG
	for win.GetMessage(&msg, 0, 0, 0) != 0 {
		win.TranslateMessage(&msg)
		win.DispatchMessage(&msg)
	}
}
