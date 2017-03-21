
err_dict={
    1 : 'Se encontraron elementos dentro de una materia antes del nombre de la materia en el .txt!\nAgrega un nombre o cambia el nombre de la materia.',
    2 : 'No se encontro el horario donde se esperaba',
    3 : 'No se encontro la barrita separadora (-) en los horarios o se encontro lejos de lo esperado\nAsegurarse de seguir el formato de SGA',
    4 : 'Error en entender el archivo.\nVerifique que tenga formato SGA con el caracter "/" al final de cada comision! ',
    5 : 'Se encontro una materia repetida? Puede ser?\nLa mejor forma de copiar materias es seguir el formato SGA\n El codigo viene primero, seguido por un hyphen (-) y despues el nombre de la materia',
    6 : 'Undefined',
    7 : 'Undefined',
}

def startread(filename):
    """Read file line by line, return info on materias"""
    
    listation_nom_materias=[]
    listation_cod_materias=[]
    matlist=[]
    ktuples= (
        'comisi','doming', 'lunes ',
        'lunes','martes','mierco',
        'mircol','jueves','vierne',
        'sabado','sbado ','sbado'   
    )
    
    found_mat=0
    mat_count=0
    error=0
    materias_info=[] #Va contener TODAS las materias y sus horarios
    with open(filename,'r') as texter:
        for line in texter: # Begins reading file line by line
            uline=line
            line=line.decode('utf8') #SGA tiene texto charset=utf8
            line=line.encode('ascii','ignore') #Para no lidiar con unicode
            l_line=line[0:6].strip()
            line=line.strip()
            if (line.strip()=='') or line[0:6]=='Comisi': #If checker
                continue
            
            elif found_mat==0 and notmat(line,ktuples):
                error=1 #No encontramos la materia antes de los horarios
                break #Possible return here!
            elif found_mat==0:
                mat_count +=1
                materia=removecode(line,mat_count)
                #Checkeemos que no se repita el codigo de la materia
                if (materia[0] not in listation_cod_materias 
                and materia[1] not in listation_nom_materias):
                    listation_cod_materias.append(materia[0])
                    listation_nom_materias.append(materia[1])
                else:
                    error=5
                found_mat+=1 #fm is 1
                continue #continuamos y agregamos horarios a esta materia
            elif found_mat==1 and line.find('-')>-1: #Para 2nda y subsecuentes materias
                mat_count+=1
                materia=removecode(line,mat_count)
                if (materia[0] not in listation_cod_materias 
                and materia[1] not in listation_nom_materias):
                    listation_cod_materias.append(materia[0])
                    listation_nom_materias.append(materia[1])
                else:
                    error=5
                continue
            elif found_mat==1 and line[0:6]!='Comisi': #Encuentra comision
                comision=line.strip()
                found_mat+=1#fm is 2
                apellidos=[]
                com_hor=[]
                continue
            elif found_mat>1 and (l_line.lower() in ktuples[1::]):
                pos_hyphen=line.find('-')
                if pos_hyphen==-1 or pos_hyphen>25:
                        error=3
                        break#Return?
                com_hor.append(line[0:pos_hyphen+7])
                found_mat+=1 #Encontro los horarios, ahora busca otros fm>2
                continue
            elif found_mat==2: 
                error=2 #Los horarios no fueron encontrados
                break #Return?
            elif found_mat>2 and line.find('/')==-1:
                comma_apellido=line.find(',')
                apellidos.append(line[0:comma_apellido]) #Agregamos Apellidos
                continue
            elif found_mat>2 and line.find('/')>-1: # Wrapup Comision
                #Mega Append
                materias_info.append([materia,comision,apellidos,com_hor])
                #Mega Append
                found_mat=1
                #Fin del circuito loop
                continue
            else:
                error=4
                continue
        else:  
            pass
            print '----------------------------'
            print 'Se genero materias_info'
            #raw_input('Continue?')
            
        if error>0:
            print err_dict[error]
    return materias_info

def notmat(texto,klines=[
        'comisi', 'vierne', 'sabado', 'lunes ', 
        'martes', 'mierco', 'jueves','doming'
        ]):
    """Return True if line belongs to materia's info, return False otherwise"""

    texto=texto.lower()
    texto=texto.strip()
    if texto[0:6] in klines:
        return True
    elif texto=='':
        return True
    else:
        return False

def removecode(texto, matcount):
    """Separate code and materia name, return as a list length=2 """

    matheria= [1, 2]
    texto=texto.strip()
    hyphen_pos=texto.find('-')
    if hyphen_pos>(-1):
        cod_mat=texto[0:hyphen_pos-1]
        mat_mat=texto[hyphen_pos+1::]
    else:
        cod_mat=matcount.strip()
        mat_mat=texto.strip()
    matheria[0]=cod_mat
    matheria[1]=mat_mat
    return matheria

def convert_horarios(horario):
    """Converts linea horarios SGA a numericos legibles
    
        Returns a five char numerical string. First digit indicates day
        the following two digits return start time, other two indicate end time
    """
    
    lunes=['lunes ','lunes']
    martes=['martes']
    miercoles=['mircol','mierco']
    jueves=['jueves']
    viernes=['vierne']
    sabado=['sabado','sbado ','sbado']
    hordia=horario[0:6].lower()
    pass
    if hordia in lunes:
        dia=1
    elif hordia in martes:
        dia=2
    elif hordia in miercoles:
        dia=3
    elif hordia in jueves:
        dia=4
    elif hordia in viernes:
        dia=5
    elif hordia in sabado:
        dia=6
    phyp=horario.find('-')
    start_time=horario[phyp-6:phyp].strip()
    end_time=horario[phyp+1::].strip()
    start_time=start_time[0:2]
    end_time=end_time[0:2]
    return str(dia)+start_time+end_time
    

def numerize_horarios(matty_info):
    """Convert day-time horarios of the mat_info to readable numeric type"""
    
    dim_mat_info=len(matty_info)
    for i in range(len(matty_info)):
        jes=matty_info[i][3]
        for j in range(len(jes)):
            jes[j]=convert_horarios(jes[j])
        matty_info[i][3]=jes
    return matty_info

def combinatrix(combos):
    """ pass  """
    dias_visualizacion= [
        #Lunes,Martes,mierc,jueves,viernes,sabado
        1   ,2     ,3      ,4     ,5      ,6, #8-9
        7   ,8     ,9      ,10    ,11     ,12,
        13  ,14    ,15     ,16    ,17     ,18,
        19  ,20    ,21     ,22    ,23     ,24,#11-12
        25  ,26    ,27     ,28    ,29     ,30,
        31  ,32    ,33     ,34    ,35     ,36,
        37  ,38    ,39     ,40    ,41     ,42,#14-15
        43  ,44    ,45     ,46    ,47     ,48,
        49  ,50    ,51     ,52    ,53     ,54,
        55  ,56    ,57     ,58    ,59     ,60,#17-18
        61  ,62    ,63     ,64    ,65     ,66,
        67  ,68    ,69     ,70    ,71     ,72,
        73  ,74    ,75     ,76    ,77     ,78,#20-21
        79  ,80    ,81     ,82    ,83     ,84,
    ]
    pass 
    return dias_visualizacion


print 'Combinatrix 2017 para ITBA \nWhittiLeaks and Co.'
print 'Click enter si existe materias.txt en el folder'
feele=''#raw_input('Caso contrario, ingresar nombre del archivo: ')
if feele.strip()=='':
    key_unref=startread('materias.txt') 
else:
    key_unref=startread(feele)
#raw_input('Quiere normalizar horarios?')
key=numerize_horarios(key_unref)
print key
raw_input('CATCH!')