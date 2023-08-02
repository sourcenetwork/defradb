use std::collections::HashMap;
use std::sync::RwLock;
use std::error::Error;
use std::{fmt, error};
use serde::Deserialize;

#[derive(Clone, PartialEq, Eq, PartialOrd, Ord, Debug, Hash)]
enum ModuleError {
    ParametersNotSetError,
}

impl error::Error for ModuleError { }

impl fmt::Display for ModuleError {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match &*self {
            ModuleError::ParametersNotSetError => f.write_str("Parameters have not been set."),
        }
    }
}

#[derive(Deserialize, Clone)]
pub struct Parameters {
    pub dst: String,
    pub value: serde_json::Value,
}

static PARAMETERS: RwLock<Option<Parameters>> = RwLock::new(None);

#[no_mangle]
pub extern fn alloc(size: usize) -> *mut u8 {
    lens_sdk::alloc(size)
}

#[no_mangle]
pub extern fn set_param(ptr: *mut u8) -> *mut u8 {
    match try_set_param(ptr) {
        Ok(_) => lens_sdk::nil_ptr(),
        Err(e) => lens_sdk::to_mem(lens_sdk::ERROR_TYPE_ID, &e.to_string().as_bytes())
    }
}

fn try_set_param(ptr: *mut u8) -> Result<(), Box<dyn Error>> {
    let parameter = lens_sdk::try_from_mem::<Parameters>(ptr)?
        .ok_or(ModuleError::ParametersNotSetError)?;

    let mut dst = PARAMETERS.write()?;
    *dst = Some(parameter);
    Ok(())
}

#[no_mangle]
pub extern fn transform(ptr: *mut u8) -> *mut u8 {
    match try_transform(ptr) {
        Ok(o) => match o {
            Some(result_json) => lens_sdk::to_mem(lens_sdk::JSON_TYPE_ID, &result_json),
            None => lens_sdk::nil_ptr(),
        },
        Err(e) => lens_sdk::to_mem(lens_sdk::ERROR_TYPE_ID, &e.to_string().as_bytes())
    }
}

fn try_transform(ptr: *mut u8) -> Result<Option<Vec<u8>>, Box<dyn Error>> {
    let mut input = match lens_sdk::try_from_mem::<HashMap<String, serde_json::Value>>(ptr)? {
        Some(v) => v,
        // Implementations of `transform` are free to handle nil however they like. In this
        // implementation we chose to return nil given a nil input.
        None => return Ok(None),
    };

    let params = PARAMETERS.read()?
        .clone()
        .ok_or(ModuleError::ParametersNotSetError)?
        .clone();

    input.insert(params.dst, params.value);

    let result_json = serde_json::to_vec(&input.clone())?;
    Ok(Some(result_json))
}

#[no_mangle]
pub extern fn inverse(ptr: *mut u8) -> *mut u8 {
    match try_inverse(ptr) {
        Ok(o) => match o {
            Some(result_json) => lens_sdk::to_mem(lens_sdk::JSON_TYPE_ID, &result_json),
            None => lens_sdk::nil_ptr(),
        },
        Err(e) => lens_sdk::to_mem(lens_sdk::ERROR_TYPE_ID, &e.to_string().as_bytes())
    }
}

fn try_inverse(ptr: *mut u8) -> Result<Option<Vec<u8>>, Box<dyn Error>> {
    let mut input = match lens_sdk::try_from_mem::<HashMap<String, serde_json::Value>>(ptr)? {
        Some(v) => v,
        // Implementations of `transform` are free to handle nil however they like. In this
        // implementation we chose to return nil given a nil input.
        None => return Ok(None),
    };

    let params = PARAMETERS.read()?
        .clone()
        .ok_or(ModuleError::ParametersNotSetError)?
        .clone();

    input.remove(&params.dst);

    let result_json = serde_json::to_vec(&input.clone())?;
    Ok(Some(result_json))
}