import idaapi
import idautils
skip_func = ['__libc_csu_init', 
             '__libc_csu_fini', 
             '_fini',
             '__do_global_dtors_aux',
             '_start',
             '_init',
             'sub_1034']

min_ea = idaapi.cvar.inf.min_ea
max_ea = idaapi.cvar.inf.max_ea
bv = idaapi.get_bytes(idaapi.get_fileregion_offset(min_ea), max_ea - min_ea)
base = idaapi.get_segm_by_name(".text").start_ea

for func_ea in idautils.Functions(base,  idaapi.get_segm_by_name(".text").end_ea):
    func_name = idaapi.get_func_name(func_ea)
    if func_name in skip_func:
        continue
    func=idaapi.get_func(func_ea)
    if func_name is None or func.flags & idaapi.FUNC_LIB:
        continue
    output = ""
    for bb in idaapi.FlowChart(idaapi.get_func(func_ea)):
        output += "0x{:x} ".format(bb.start_ea - base)
    print(output)