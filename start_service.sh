datetime=$(date +%Y%m%d_%H%M%S)
log_file="service_${datetime}.log"

nohup ./serv -app=service -log=console > ${log_file} 2>&1 &