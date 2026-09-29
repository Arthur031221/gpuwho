// rusage_probe.swift
//
// Throwaway probe used while researching gpuwho's technical bet: does
// proc_pid_rusage expose a per-process GPU time counter on this machine
// (macOS 26.6, Apple M5, Xcode SDK MacOSX26.5/27.0)?
//
// The brief asked to check for a field named ri_gpu_time_ns on
// rusage_info_v4 or later. This probe calls proc_pid_rusage with the
// newest available flavor (RUSAGE_INFO_V6) for a given PID and prints
// every field that exists in the struct on this SDK.
//
// Result: rusage_info_v4, v5 and v6 (checked against
// /Library/Developer/CommandLineTools/SDKs/MacOSX26.5.sdk/usr/include/sys/resource.h
// and the MacOSX27.0.sdk and the private Kernel.framework header) have no
// GPU time field at all, public or private. v6 adds ri_neural_footprint,
// ri_lifetime_max_neural_footprint and ri_interval_max_neural_footprint,
// which are Neural Engine *memory* counters (bytes), not time or percent.
// There is no ri_gpu_time_ns anywhere in the Darwin headers on this system.
//
// Build:  swiftc rusage_probe.swift -o rusage_probe
// Run:    ./rusage_probe <pid>
//
// This file is kept for reference. gpuwho itself does not use cgo or a
// Swift helper: see README.md "How it works" for what it uses instead.

import Darwin

let RUSAGE_INFO_V6: Int32 = 6

struct rusage_info_v6 {
    var ri_uuid: (UInt8, UInt8, UInt8, UInt8, UInt8, UInt8, UInt8, UInt8,
                  UInt8, UInt8, UInt8, UInt8, UInt8, UInt8, UInt8, UInt8) =
        (0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0)
    var ri_user_time: UInt64 = 0
    var ri_system_time: UInt64 = 0
    var ri_pkg_idle_wkups: UInt64 = 0
    var ri_interrupt_wkups: UInt64 = 0
    var ri_pageins: UInt64 = 0
    var ri_wired_size: UInt64 = 0
    var ri_resident_size: UInt64 = 0
    var ri_phys_footprint: UInt64 = 0
    var ri_proc_start_abstime: UInt64 = 0
    var ri_proc_exit_abstime: UInt64 = 0
    var ri_child_user_time: UInt64 = 0
    var ri_child_system_time: UInt64 = 0
    var ri_child_pkg_idle_wkups: UInt64 = 0
    var ri_child_interrupt_wkups: UInt64 = 0
    var ri_child_pageins: UInt64 = 0
    var ri_child_elapsed_abstime: UInt64 = 0
    var ri_diskio_bytesread: UInt64 = 0
    var ri_diskio_byteswritten: UInt64 = 0
    var ri_cpu_time_qos_default: UInt64 = 0
    var ri_cpu_time_qos_maintenance: UInt64 = 0
    var ri_cpu_time_qos_background: UInt64 = 0
    var ri_cpu_time_qos_utility: UInt64 = 0
    var ri_cpu_time_qos_legacy: UInt64 = 0
    var ri_cpu_time_qos_user_initiated: UInt64 = 0
    var ri_cpu_time_qos_user_interactive: UInt64 = 0
    var ri_billed_system_time: UInt64 = 0
    var ri_serviced_system_time: UInt64 = 0
    var ri_logical_writes: UInt64 = 0
    var ri_lifetime_max_phys_footprint: UInt64 = 0
    var ri_instructions: UInt64 = 0
    var ri_cycles: UInt64 = 0
    var ri_billed_energy: UInt64 = 0
    var ri_serviced_energy: UInt64 = 0
    var ri_interval_max_phys_footprint: UInt64 = 0
    var ri_runnable_time: UInt64 = 0
    var ri_flags: UInt64 = 0
    var ri_user_ptime: UInt64 = 0
    var ri_system_ptime: UInt64 = 0
    var ri_pinstructions: UInt64 = 0
    var ri_pcycles: UInt64 = 0
    var ri_energy_nj: UInt64 = 0
    var ri_penergy_nj: UInt64 = 0
    var ri_secure_time_in_system: UInt64 = 0
    var ri_secure_ptime_in_system: UInt64 = 0
    var ri_neural_footprint: UInt64 = 0
    var ri_lifetime_max_neural_footprint: UInt64 = 0
    var ri_interval_max_neural_footprint: UInt64 = 0
}

@_silgen_name("proc_pid_rusage")
func proc_pid_rusage(_ pid: Int32, _ flavor: Int32, _ buffer: UnsafeMutableRawPointer?) -> Int32

let args = CommandLine.arguments
guard args.count > 1, let pid = Int32(args[1]) else {
    print("usage: rusage_probe <pid>")
    exit(1)
}

var info = rusage_info_v6()
let rc = withUnsafeMutablePointer(to: &info) { ptr -> Int32 in
    return proc_pid_rusage(pid, RUSAGE_INFO_V6, UnsafeMutableRawPointer(ptr))
}

if rc != 0 {
    perror("proc_pid_rusage")
    exit(1)
}

print("pid \(pid) rusage_info_v6 (RUSAGE_INFO_CURRENT):")
print("  ri_user_time            = \(info.ri_user_time) ns   (CPU, not GPU)")
print("  ri_system_time          = \(info.ri_system_time) ns")
print("  ri_phys_footprint       = \(info.ri_phys_footprint) bytes")
print("  ri_instructions         = \(info.ri_instructions)")
print("  ri_cycles               = \(info.ri_cycles)")
print("  ri_energy_nj            = \(info.ri_energy_nj) nJ  (package energy, not GPU-scoped)")
print("  ri_neural_footprint     = \(info.ri_neural_footprint) bytes (ANE memory, not time)")
print("  -- there is no ri_gpu_time_ns field in this struct on this SDK --")
