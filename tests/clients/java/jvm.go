// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

//go:build javaclient

package java

/*
#include <stdlib.h>
#include <jni.h>
#include "../../../cbindings/defra_structs.h"
#include "jvmboot.h"
#include "jnicall.h"

extern int defra_register_node_natives(JNIEnv* env, jclass cls, char* errbuf, int errbufLen);
extern int defra_register_transaction_natives(JNIEnv* env, jclass cls, char* errbuf, int errbufLen);

// Accessors onto registernatives.c's own JNINativeMethod tables
// Having these allows us to not have to duplicate the table into this file

extern int defra_node_native_method_count(void);
extern const char* defra_node_native_method_name(int i);
extern const char* defra_node_native_method_signature(int i);
extern int defra_transaction_native_method_count(void);
extern const char* defra_transaction_native_method_name(int i);
extern const char* defra_transaction_native_method_signature(int i);
*/
import "C"

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"unsafe"
)

// This file contains functionality related to the JVM's lifecycle, including its setup, and
// the functions to create nodes and transactions inside it.

const (
	javaHomeEnvName       = "JAVA_HOME"              // Path to the Java JDK
	javaJarEnvName        = "DEFRA_JAVA_JAR"         // Path to the Defra Java SDK Jar file
	javaJVMOptsEnvName    = "DEFRA_JAVA_JVM_OPTS"    // Optional command-line JVM options, split on whitespace
	javaWrapperDirEnvName = "DEFRA_JAVA_WRAPPER_DIR" // Overrides where defaultJarPath looks for the built jar; must match tools/scripts/build-java-client.sh's own WRAPPER_DIR
)

var (
	// Ensures the one-time JVM setup (setupJVM) only ever runs once, caching its result
	setupOnce sync.Once
	setupErr  error

	// The global JVM reference
	javaVM *C.JavaVM

	// The global DefraNode class, constructor, and field IDs
	nodeClass       C.jclass
	nodeConstructor C.jmethodID
	nodeMethodIDs   map[string]C.jmethodID

	// The global DefraTransaction class, constructor, and array of method IDs
	transactionClass       C.jclass
	transactionConstructor C.jmethodID
	transactionMethodIDs   map[string]C.jmethodID

	// The global DefraResult class, constructor, and field IDs
	resultClass       C.jclass
	resultStatusField C.jfieldID
	resultErrorField  C.jfieldID
	resultValueField  C.jfieldID

	// The global DefraTransactionResult class, constructor, and array of method IDs
	txnResultClass       C.jclass
	txnResultStatusField C.jfieldID
	txnResultErrorField  C.jfieldID
	txnResultPtrField    C.jfieldID
)

// ensureJVM calls setupJVM in such a way that it only ever happens once, no matter how many times, or from how many
// goroutines it gets called.
func ensureJVM() error {
	setupOnce.Do(func() {
		setupErr = setupJVM()
	})
	return setupErr
}

// setupJVM does everything necessary to get the embedded JVM ready for use. It gets called lazily, via ensureJVM,
// the first time any goroutine actually needs to make a JNI call.
func setupJVM() error {
	// Utilize environment variables to get the Java JDK and Defra Jar paths
	javaHome := os.Getenv(javaHomeEnvName)
	if javaHome == "" {
		return fmt.Errorf(errFmtJavaHomeNotSet, javaHomeEnvName)
	}
	jarPath := os.Getenv(javaJarEnvName)
	// If there is no Jar path, we can see if one exists in the build directory
	if jarPath == "" {
		jarPath = defaultJarPath()
	}
	if jarPath == "" {
		// If there is no Jar there either, we can't continue
		return fmt.Errorf(errFmtJarNotSet, javaJarEnvName)
	}
	jvmLib, err := jvmLibraryPath(javaHome)
	if err != nil {
		return err
	}

	// Convert environment variables to C strings that can be used
	cJvmLib := C.CString(jvmLib)
	defer C.free(unsafe.Pointer(cJvmLib))
	cJar := C.CString(jarPath)
	defer C.free(unsafe.Pointer(cJar))
	cExtraOpts := C.CString(os.Getenv(javaJVMOptsEnvName)) // This is optional, and might be blank
	defer C.free(unsafe.Pointer(cExtraOpts))

	// Try to start the JVM
	var vm *C.JavaVM
	var env *C.JNIEnv
	var errbuf [C.DEFRA_ERRBUF_LEN]C.char
	if retcode := C.defra_start_jvm(cJvmLib, cJar, cExtraOpts, &vm, &env, &errbuf[0], C.int(len(errbuf))); retcode != 0 {
		return fmt.Errorf(errFmtStartJVMFailed, jvmLib, C.GoString(&errbuf[0]))
	}
	javaVM = vm

	// Create a helper function for getting a class, and an error buffer that can be used for it
	var errbuf2 [C.DEFRA_ERRBUF_LEN]C.char
	getClass := func(name string) (C.jclass, error) {
		cName := C.CString(name)
		defer C.free(unsafe.Pointer(cName))
		cls := C.defra_find_global_class(env, cName, &errbuf2[0], C.int(len(errbuf2)))
		if cls == 0 {
			return 0, fmt.Errorf(errFmtJNIError, C.GoString(&errbuf2[0]))
		}
		return cls, nil
	}

	// Use the helper function to get the global DefraNode, DefraResult, DefraTransaction,
	// and DefraTransactionResult classes
	var err2 error
	if nodeClass, err2 = getClass("source/defra/DefraNode"); err2 != nil {
		return err2
	}
	if resultClass, err2 = getClass("source/defra/DefraResult"); err2 != nil {
		return err2
	}
	if transactionClass, err2 = getClass("source/defra/DefraTransaction"); err2 != nil {
		return err2
	}
	if txnResultClass, err2 = getClass("source/defra/DefraTransactionResult"); err2 != nil {
		return err2
	}

	// Register the native functions
	// The TL;DR is that this allows us to use the implementation of the exported C functions that are already compiled into,
	// and available to the test binary. For more information, see defra_register_natives inside jnicall.c
	if retcode := C.defra_register_node_natives(env, nodeClass, &errbuf2[0], C.int(len(errbuf2))); retcode != 0 {
		return fmt.Errorf(errFmtRegisterNodeNativesFailed, C.GoString(&errbuf2[0]))
	}

	if retcode := C.defra_register_transaction_natives(env, transactionClass, &errbuf2[0], C.int(len(errbuf2))); retcode != 0 {
		return fmt.Errorf(errFmtRegisterTransactionNativesFailed, C.GoString(&errbuf2[0]))
	}

	// Create a helper function that gets a method with a given name belonging to a class
	getMethod := func(cls C.jclass, name, sig string) (C.jmethodID, error) {
		cName := C.CString(name)
		defer C.free(unsafe.Pointer(cName))
		cSig := C.CString(sig)
		defer C.free(unsafe.Pointer(cSig))
		m := C.defra_get_method_id(env, cls, cName, cSig, &errbuf2[0], C.int(len(errbuf2)))
		if m == nil {
			return nil, fmt.Errorf(errFmtJNIError, C.GoString(&errbuf2[0]))
		}
		return m, nil
	}

	// Create a helper function that gets a field with a given name belonging to a class
	getField := func(cls C.jclass, name, sig string) (C.jfieldID, error) {
		cName := C.CString(name)
		defer C.free(unsafe.Pointer(cName))
		cSig := C.CString(sig)
		defer C.free(unsafe.Pointer(cSig))
		f := C.defra_get_field_id(env, cls, cName, cSig, &errbuf2[0], C.int(len(errbuf2)))
		if f == nil {
			return nil, fmt.Errorf(errFmtJNIError, C.GoString(&errbuf2[0]))
		}
		return f, nil
	}

	// Get the DefraNode's constructor method
	if nodeConstructor, err2 = getMethod(nodeClass, "<init>", "(J)V"); err2 != nil {
		return err2
	}

	// For each of the DefraNode's methods, get a method ID, and add it to the nodeMethodIDs map
	nodeMethodCount := int(C.defra_node_native_method_count())
	nodeMethodIDs = make(map[string]C.jmethodID, nodeMethodCount)
	for i := 0; i < nodeMethodCount; i++ {
		name := C.GoString(C.defra_node_native_method_name(C.int(i)))
		sig := C.GoString(C.defra_node_native_method_signature(C.int(i)))
		mid, err3 := getMethod(nodeClass, name, sig)
		if err3 != nil {
			return err3
		}
		nodeMethodIDs[name] = mid
	}

	// Get the DefraTransaction constructor method
	if transactionConstructor, err2 = getMethod(transactionClass, "<init>", "(J)V"); err2 != nil {
		return err2
	}

	// For each of the DefraTransaction's methods, get a method ID, and add it to the transactionMethodIDs map
	transactionMethodCount := int(C.defra_transaction_native_method_count())
	transactionMethodIDs = make(map[string]C.jmethodID, transactionMethodCount)
	for i := 0; i < transactionMethodCount; i++ {
		name := C.GoString(C.defra_transaction_native_method_name(C.int(i)))
		sig := C.GoString(C.defra_transaction_native_method_signature(C.int(i)))
		mid, err3 := getMethod(transactionClass, name, sig)
		if err3 != nil {
			return err3
		}
		transactionMethodIDs[name] = mid
	}

	// Set the remaining global references by getting all the appropriate fields by name,
	// on the DefraResult and DefraTransactionResult classes
	if resultStatusField, err2 = getField(resultClass, "status", "I"); err2 != nil {
		return err2
	}
	if resultErrorField, err2 = getField(resultClass, "error", "Ljava/lang/String;"); err2 != nil {
		return err2
	}
	if resultValueField, err2 = getField(resultClass, "value", "Ljava/lang/String;"); err2 != nil {
		return err2
	}
	if txnResultStatusField, err2 = getField(txnResultClass, "status", "I"); err2 != nil {
		return err2
	}
	if txnResultErrorField, err2 = getField(txnResultClass, "error", "Ljava/lang/String;"); err2 != nil {
		return err2
	}
	if txnResultPtrField, err2 = getField(txnResultClass, "txnPtr", "J"); err2 != nil {
		return err2
	}

	return nil
}

// jvmLibraryPath gets the JVM library file with the appropriate name, based on the OS.
//
// JDK 9 (JEP 220) flattened the jre/ subdirectory out of the JDK image; JDK 8 and earlier still
// nest the JVM library under jre/, and on Linux additionally nest an arch-named directory under
// jre/lib/. The JDK 9+ layout is tried first, falling back to the JDK 8 layout if it doesn't exist.
func jvmLibraryPath(javaHome string) (string, error) {
	var primary, fallback string
	switch runtime.GOOS {
	case "windows":
		primary = filepath.Join(javaHome, "bin", "server", "jvm.dll")
		fallback = filepath.Join(javaHome, "jre", "bin", "server", "jvm.dll")
	case "darwin":
		primary = filepath.Join(javaHome, "lib", "server", "libjvm.dylib")
		fallback = filepath.Join(javaHome, "jre", "lib", "server", "libjvm.dylib")
	case "linux":
		primary = filepath.Join(javaHome, "lib", "server", "libjvm.so")
		fallback = filepath.Join(javaHome, "jre", "lib", jdk8LinuxArch(), "server", "libjvm.so")
	default:
		return "", fmt.Errorf(errFmtUnsupportedGOOS, runtime.GOOS)
	}

	if fileExists(primary) {
		return primary, nil
	}
	if fileExists(fallback) {
		return fallback, nil
	}
	// If neither exists, there will be a dlopen failure downstream. But it may make more sense
	// to use the modern path rather than the legacy one.
	return primary, nil
}

// fileExists reports whether path can be stat'd - used by jvmLibraryPath to probe for the JVM
// library without distinguishing why a candidate might be missing (permissions, not-a-JDK, etc.);
// any of those should fall through to the next candidate, or the final default, the same way.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// jdk8LinuxArch maps runtime.GOARCH to the directory name thatJDK 8's Linux layout places libjvm.so:
// jre/lib/<arch>/server/a
func jdk8LinuxArch() string {
	switch runtime.GOARCH {
	case "amd64":
		return "amd64"
	case "386":
		return "i386"
	case "arm64":
		return "aarch64"
	default:
		return runtime.GOARCH
	}
}

// defaultJarPath returns the jar path produced by `make build-java-client` if it exists.
// Returns "" if no such repo root or no jar file can be found.
//
// If DEFRA_JAVA_WRAPPER_DIR is set, the jar is looked for under that directory instead of the
// standard .javaclient/defradb-java-sdk - this must match tools/scripts/build-java-client.sh's
// own WRAPPER_DIR default, since that script is what produces the jar in the first place.
func defaultJarPath() string {
	if wrapperDir := os.Getenv(javaWrapperDirEnvName); wrapperDir != "" {
		jar := filepath.Join(wrapperDir, "build", "libs", "defradb.jar")
		if _, err := os.Stat(jar); err == nil {
			return jar
		}
		return ""
	}

	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			jar := filepath.Join(dir, ".javaclient", "defradb-java-sdk", "build", "libs", "defradb.jar")
			if _, err := os.Stat(jar); err == nil {
				return jar
			}
			return ""
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// attach attaches the calling OS thread to the embedded JVM, locking the current goroutine to that
// thread for the duration (a JNIEnv is only valid on the thread that obtained it). The returned detach
// function must be called before the goroutine does anything that might yield the thread back to the scheduler.
func attach() (*C.JNIEnv, func(), error) {
	runtime.LockOSThread()
	if err := ensureJVM(); err != nil {
		runtime.UnlockOSThread()
		return nil, nil, err
	}
	var env *C.JNIEnv
	var errbuf [C.DEFRA_ERRBUF_LEN]C.char
	if retcode := C.defra_attach_thread(javaVM, &env, &errbuf[0], C.int(len(errbuf))); retcode != 0 {
		runtime.UnlockOSThread()
		return nil, nil, fmt.Errorf(errFmtJNIError, C.GoString(&errbuf[0]))
	}
	detach := func() {
		C.defra_detach_thread(javaVM)
		runtime.UnlockOSThread()
	}
	return env, detach, nil
}

// newNodeObject creates the DefraNode Java object wrapping an already-running *node.Node, given the
// uintptr value of a cgo.Handle for that node. It returns a global reference.
func newNodeObject(nodePtr uintptr) (C.jobject, error) {
	// Pin this goroutine to its OS thread.
	// Doing this prevents a SIGSEGV crash, because a JNIEnv is only valid on the thread that obtained it.
	env, detach, err := attach()
	if err != nil {
		return 0, err
	}
	defer detach()

	// Make the JNI call to create the node
	args := []C.DefraArg{{kind: C.DEFRA_ARG_LONG, j: C.jlong(nodePtr)}}
	var errbuf [C.DEFRA_ERRBUF_LEN]C.char
	obj := C.defra_new_object(env, nodeClass, nodeConstructor, &args[0], 1, &errbuf[0], C.int(len(errbuf)))
	if obj == 0 {
		return 0, fmt.Errorf(errFmtCreateNodeFailed, C.GoString(&errbuf[0]))
	}
	return obj, nil
}

// newTransactionObject creates the DefraTransaction Java object wrapping an already-running transaction,
// given the uintptr value of the cgo.Handle returned by createTransactionWithHandle. It returns a global reference.
func newTransactionObject(txnPtr uintptr) (C.jobject, error) {
	// Pin this goroutine to its OS thread.
	// Doing this prevents a SIGSEGV crash, because a JNIEnv is only valid on the thread that obtained it.
	env, detach, err := attach()
	if err != nil {
		return 0, err
	}
	defer detach()

	// Make the JNI call to create the transaction
	args := []C.DefraArg{{kind: C.DEFRA_ARG_LONG, j: C.jlong(txnPtr)}}
	var errbuf [C.DEFRA_ERRBUF_LEN]C.char
	obj := C.defra_new_object(env, transactionClass, transactionConstructor, &args[0], 1, &errbuf[0], C.int(len(errbuf)))
	if obj == 0 {
		return 0, fmt.Errorf(errFmtCreateTransactionFailed, C.GoString(&errbuf[0]))
	}
	return obj, nil
}
