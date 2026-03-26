#include <stdlib.h>
#include <jni.h>

// These must be set by the Java side before any BLE operations are performed
// The wrapper layer provides a function to register a BLE interface, which will set these:
//
// JNIEXPORT void JNICALL
// Java_source_defra_DefraBLE_registerBleInterface(JNIEnv *env, jclass clazz, jobject bleInterface) {
//    (*env)->GetJavaVM(env, &gJVM);
//    gBleInterface = (*env)->NewGlobalRef(env, bleInterface);
//}
//
// The user would be responsible for calling this function before using any BLE operations.
//
extern JavaVM* gJVM;
extern jobject gBleInterface;

// The following functions allow our Go code BLE driver to call into the Java side
// to handle BLE events that occur there. These would be implemented in a package that
// exists as part of the user's Android project.

jboolean CallDialPeer(const char* remotePID) {
    JNIEnv* env;
    (*gJVM)->AttachCurrentThread(gJVM, &env, NULL);
    jclass cls = (*env)->GetObjectClass(env, gBleInterface);
    jmethodID mid = (*env)->GetMethodID(env, cls, "dialPeer", "(Ljava/lang/String;)Z");
    jstring jPID = (*env)->NewStringUTF(env, remotePID);
    jboolean result = (*env)->CallBooleanMethod(env, gBleInterface, mid, jPID);
    (*env)->DeleteLocalRef(env, jPID);
    return result;
}

jboolean CallSendToPeer(const char* remotePID, void* payload, int length) {
    JNIEnv* env;
    (*gJVM)->AttachCurrentThread(gJVM, &env, NULL);
    jclass cls = (*env)->GetObjectClass(env, gBleInterface);
    jmethodID mid = (*env)->GetMethodID(env, cls, "sendToPeer", "(Ljava/lang/String;[B)Z");
    jstring jPID = (*env)->NewStringUTF(env, remotePID);
    jbyteArray jPayload = (*env)->NewByteArray(env, length);
    (*env)->SetByteArrayRegion(env, jPayload, 0, length, (jbyte*)payload);
    jboolean result = (*env)->CallBooleanMethod(env, gBleInterface, mid, jPID, jPayload);
    (*env)->DeleteLocalRef(env, jPID);
    (*env)->DeleteLocalRef(env, jPayload);
    return result;
}

void CallCloseConnWithPeer(const char* remotePID) {
    JNIEnv* env;
    (*gJVM)->AttachCurrentThread(gJVM, &env, NULL);
    jclass cls = (*env)->GetObjectClass(env, gBleInterface);
    jmethodID mid = (*env)->GetMethodID(env, cls, "closeConnWithPeer", "(Ljava/lang/String;)V");
    jstring jPID = (*env)->NewStringUTF(env, remotePID);
    (*env)->CallVoidMethod(env, gBleInterface, mid, jPID);
    (*env)->DeleteLocalRef(env, jPID);
}