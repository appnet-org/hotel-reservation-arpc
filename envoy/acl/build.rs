const PROTO: &str = "../user-proto/user.proto";
fn main() -> Result<(), Box<dyn std::error::Error>> {
    println!("cargo:rerun-if-changed={PROTO}");
    prost_build::compile_protos(&[PROTO], &["../user-proto"])?;
    Ok(())
}
