// trackpad_dump2.c  ← NEW!
#include <ApplicationServices/ApplicationServices.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdio.h>

static const char* phaseName(int64_t p){
    switch(p){case 128:return "MayBegin";case 1:return "Began";
              case 2:return "Changed";case 4:return "Ended";
              case 8:return "Cancelled";default:return "?";}
}
static const char* subtypeName(int64_t s){
    switch(s){
        case 2:  return "Pinch";          // 2-finger pinch/rotate
        case 6:  return "Scroll";         // 2-finger scroll
        case 15: return "Swipe(legacy)";  // Mojave 以前
        case 23: return "Swipe";          // 3/4-finger, Sonoma+
        default: return "?";
    }
}
static void dumpMask(int64_t m){
    if(!m){printf("mask=0 ");return;}
    printf("mask=0x%llX [",m);
    // mask値はフェーズを表す（方向ではない）
    if(m&0x1)printf("Began ");
    if(m&0x2)printf("Changed ");
    if(m&0x4)printf("0x4 ");  // 不明な値
    if(m&0x8)printf("Ended/Cancelled ");
    printf("] ");
}

static CGEventRef cb(CGEventTapProxy proxy,CGEventType t,CGEventRef e,void* info){
    double ts = CGEventGetTimestamp(e)/1e9;
    int64_t sub = CGEventGetIntegerValueField(e,110);
    int64_t ph  = CGEventGetIntegerValueField(e,132);
    int64_t msk = CGEventGetIntegerValueField(e,134);
    int64_t flg = CGEventGetIntegerValueField(e,115);

    if(t==kCGEventScrollWheel){
        double dx = CGEventGetDoubleValueField(e,96);
        double dy = CGEventGetDoubleValueField(e,97);
        if(!CGEventGetIntegerValueField(e,kCGScrollWheelEventIsContinuous)) return e;
        printf("[%.3f] Scroll %s Δ(%.1f,%.1f)\n",ts,phaseName(ph),dx,dy);
        return e;
    }
    if(t==29||t==30||t==31){
        double gdx = CGEventGetDoubleValueField(e,116);
        double gdy = CGEventGetDoubleValueField(e,119);
        printf("[%.3f] Gesture type=%d(%s) sub=%lld(%s) phase=%s ",
               ts,t,(t==29?"Begin":t==30?"Change":"End"),
               sub,subtypeName(sub),phaseName(ph));
        dumpMask(msk);
        printf("flags=0x%llX\n",flg);
        if(gdx||gdy) printf("          Δ(%.1f,%.1f)\n",gdx,gdy);
    }
    return e;
}
int main(){
    puts("=== CGEvent dump v2 ===");
    CGEventMask m = CGEventMaskBit(kCGEventScrollWheel)
                  | CGEventMaskBit(29)|CGEventMaskBit(30)|CGEventMaskBit(31);
    CFMachPortRef tap = CGEventTapCreate(kCGSessionEventTap,
                                         kCGHeadInsertEventTap,
                                         kCGEventTapOptionListenOnly,
                                         m,cb,NULL);
    if(!tap){fputs("Need Accessibility permission\n",stderr);return 1;}
    CFRunLoopSourceRef src=CFMachPortCreateRunLoopSource(kCFAllocatorDefault,tap,0);
    CFRunLoopAddSource(CFRunLoopGetCurrent(),src,kCFRunLoopCommonModes);
    CGEventTapEnable(tap,true);
    puts("Listening…  Ctrl+C to quit");
    CFRunLoopRun();
    return 0;
}