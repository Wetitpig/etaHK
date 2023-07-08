use std::ffi::CString;

mod ctb_nwfb;

#[no_mangle]
pub extern "C" fn bus_crawl() -> *const libc::c_char {
	let final_file = CString::new("Hello").expect("CString::new failed");
	final_file.into_raw()
}
