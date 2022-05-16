datetime=$(date +%Y%m%d_%H%M%S)
log_file="client_${datetime}.log"

nohup ./serv -app=client -cli_num=1200 -log=console > ${log_file} 2>&1 &