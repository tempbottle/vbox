#include "VBoxCAPIGlue.h"

// Wrappers declared in vbox.c
HRESULT GoVboxFAILED(HRESULT result);
HRESULT GoVboxArrayOutFree(void* array);
void GoVboxUtf8Free(char* cstring);


HRESULT GoVboxGetMouseAbsoluteSupported(IMouse* cmouse, PRBool* csupported) {
  return IMouse_GetAbsoluteSupported(cmouse, csupported);
}
HRESULT GoVboxGetMouseRelativeSupported(IMouse* cmouse, PRBool* csupported) {
  return IMouse_GetRelativeSupported(cmouse, csupported);
}
HRESULT GoVboxPutMouseEvent(IMouse* cmouse, PRInt32 cdx, PRInt32 cdy,
    PRInt32 cdz, PRInt32 cdw, PRInt32 cbuttonState) {
  return IMouse_PutMouseEvent(cmouse, cdx, cdy, cdz, cdw, cbuttonState);
}
HRESULT GoVboxPutMouseEventAbsolute(IMouse* cmouse, PRInt32 cx, PRInt32 cy,
    PRInt32 cdz, PRInt32 cdw, PRInt32 cbuttonState) {
  return IMouse_PutMouseEventAbsolute(cmouse, cx, cy, cdz, cdw, cbuttonState);
}
HRESULT GoVboxIMouseRelease(IMouse* cmouse) {
  return IMouse_Release(cmouse);
}
